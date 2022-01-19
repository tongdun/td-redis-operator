package svc

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"time"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"td-redis-operator/pkg/apis/cache/v1alpha1"
	endpointsutil "td-redis-operator/third_party/kubernetes/pkg/api/v1/endpoints"
	podutil "td-redis-operator/third_party/kubernetes/pkg/api/v1/pod"
)

func (c *Controller) syncService(key string) error {
	startTime := time.Now()

	defer func() {
		klog.V(4).Infof("Finished syncing service %q endpoints. (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	svc, err := c.svcLister.Services(ns).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// Delete the corresponding endpoint, as the service has been deleted.
		// TODO: Please note that this will delete an endpoint when a
		// service is deleted. However, if we're down at the time when
		// the service is deleted, we will miss that deletion, so this
		// doesn't completely solve the problem. See #6877.
		if err := c.kubeClient.CoreV1().Endpoints(ns).Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}

		c.triggerTimeTracker.DeleteService(ns, name)

		return nil
	}

	selectorAnno, ok := svc.Annotations[v1alpha1.RedisClusterServiceSelectorAnnotation]
	if !ok {
		// not service should be handled by this controller
		return nil
	}

	/*statusAnno, ok := svc.Annotations[v1alpha1.SingletonServiceStatusAnnotation]
	if ok && statusAnno == v1alpha1.SingletonServiceDisabled {
		if err := c.kubeClient.CoreV1().Endpoints(ns).Delete(context.TODO(),name, metav1.DeleteOptions{}); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}
		return nil
	}*/

	selector, err := parseSelector(selectorAnno)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Error converting svc annotation selector to selector: %v", err))
		return nil
	}
	pods, err := c.podLister.Pods(ns).List(selector)

	if err != nil {
		return err
	}

	endpointsLastChangeTriggerTime := c.triggerTimeTracker.ComputeEndpointLastChangeTriggerTime(ns, svc, pods)
	isCreated := true

	endpoints, err := c.endpointsLister.Endpoints(ns).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			isCreated = false
			endpoints = &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ns,
					Labels:    svc.Labels,
				},
			}
		} else {
			return err
		}
	}

	mp, err := c.redisClusterLister.RedisClusters(ns).Get(name)
	if err != nil {
		return nil
	}
	subsets, _ := genSubsetFromPods(pods, svc, mp)
	//subsets = repackSubsets(subsets, readySubset, endpoints.Subsets)
	if isCreated &&
		apiequality.Semantic.DeepEqual(endpoints.Subsets, subsets) &&
		apiequality.Semantic.DeepEqual(endpoints.Labels, svc.Labels) {
		klog.V(5).Infof("endpoints are equal for %s/%s, skipping update", svc.Namespace, svc.Name)
		return nil
	}

	cp := endpoints.DeepCopy()
	cp.Subsets = subsets
	cp.Labels = svc.Labels

	if !endpointsLastChangeTriggerTime.IsZero() {
		if cp.Annotations == nil {
			cp.Annotations = make(map[string]string)
		}

		cp.Annotations[corev1.EndpointsLastChangeTriggerTime] =
			endpointsLastChangeTriggerTime.Format(time.RFC3339Nano)
	} else {
		// No new trigger time, clear the annotation.
		delete(cp.Annotations, corev1.EndpointsLastChangeTriggerTime)
	}

	if err := c.updateEndpoints(cp, isCreated); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}

func (c *Controller) updateEndpoints(ep *corev1.Endpoints, isCreated bool) error {
	klog.V(4).Infof("Update endpoints for %v/%v", ep.Namespace, ep.Name)

	if isCreated {
		if _, err := c.kubeClient.CoreV1().Endpoints(ep.Namespace).Update(context.TODO(), ep, metav1.UpdateOptions{}); err != nil {
			c.eventRecorder.Eventf(ep,
				corev1.EventTypeWarning,
				"FailedToUpdateEndpoint",
				"Failed to update endpoint %v/%v: %v",
				ep.Namespace,
				ep.Name,
				err,
			)

			return err
		}

		return nil
	}

	if _, err := c.kubeClient.CoreV1().Endpoints(ep.Namespace).Create(context.TODO(), ep, metav1.CreateOptions{}); err != nil {
		if errors.IsForbidden(err) {
			// A request is forbidden primarily for two reasons:
			// 1. namespace is terminating, endpoint creation is not allowed by default.
			// 2. policy is misconfigured, in which case no service would function anywhere.
			// Given the frequency of 1, we log at a lower level.
			klog.V(5).Infof("Forbidden from creating endpoints: %v", err)
		}

		c.eventRecorder.Eventf(ep,
			corev1.EventTypeWarning,
			"FailedToCreateEndpoint",
			"Failed to create endpoint for service %v/%v: %v",
			ep.Namespace,
			ep.Name,
			err,
		)

		return err
	}

	return nil
}

func podInCluster(pod *corev1.Pod, mp *v1alpha1.RedisCluster) bool {
	rdb := redis.NewClient(&redis.Options{
		Addr:     pod.Status.PodIP + ":6379",
		Password: mp.Spec.Secret,
	})
	if rs, err := rdb.ClusterSlots().Result(); err != nil {
		return false
	} else {
		if len(rs) == 0 {
			return false
		}
	}
	return true
}

func addEndpointSubset(subsets []corev1.EndpointSubset,
	readySubset map[addressKey]struct{},
	pod *corev1.Pod,
	epa corev1.EndpointAddress,
	epps []corev1.EndpointPort,
	mp *v1alpha1.RedisCluster,
) ([]corev1.EndpointSubset, map[addressKey]struct{}) {
	if podutil.IsPodReady(pod) && podInCluster(pod, mp) {
		key := getAddressKey(&epa)
		readySubset[key] = struct{}{}

		subsets = append(subsets, corev1.EndpointSubset{
			Addresses: []corev1.EndpointAddress{epa},
			Ports:     epps,
		})
	} else if shouldPodBeInEndpoints(pod) {
		subsets = append(subsets, corev1.EndpointSubset{
			NotReadyAddresses: []corev1.EndpointAddress{epa},
			Ports:             epps,
		})
	}

	return subsets, readySubset
}

func genSubsetFromPods(pods []*corev1.Pod, svc *corev1.Service, mp *v1alpha1.RedisCluster) ([]corev1.EndpointSubset, map[addressKey]struct{}) {
	subsets := []corev1.EndpointSubset{}
	readySubset := map[addressKey]struct{}{}

	for _, pod := range pods {
		if len(pod.Status.PodIP) == 0 {
			klog.V(5).Infof("Failed to find an IP for pod %s/%s", pod.Namespace, pod.Name)
			continue
		}

		if pod.DeletionTimestamp != nil {
			klog.V(5).Infof("Pod is being deleted %s/%s", pod.Namespace, pod.Name)
			continue
		}

		// NOTE(bo.liub): dual stack is not considered
		// if dual stack is needed, please see k8s.io/pkg/controller/endpoint
		ep := podToEndpointAddress(pod)
		epa := *ep

		hostname := pod.Spec.Hostname
		if len(hostname) > 0 && pod.Spec.Subdomain == svc.Name && svc.Namespace == pod.Namespace {
			epa.Hostname = hostname
		}

		epps := []corev1.EndpointPort{}

		// NOTE(bo.liub): ignore headless service
		for i := range svc.Spec.Ports {
			svcPort := &svc.Spec.Ports[i]

			portNum, err := podutil.FindPort(pod, svcPort)
			if err != nil {
				klog.V(4).Infof("Failed to find port for service %s/%s: %v", svc.Namespace, svc.Name, err)
				continue
			}

			epp := corev1.EndpointPort{
				Name:     svcPort.Name,
				Port:     int32(portNum),
				Protocol: svcPort.Protocol,
			}
			epps = append(epps, epp)
		}

		if len(epps) != 0 {
			subsets, readySubset = addEndpointSubset(subsets, readySubset, pod, epa, epps, mp)
		}
	}

	return subsets, readySubset
}

func repackSubsets(subsets []corev1.EndpointSubset,
	readySubset map[addressKey]struct{},
	current []corev1.EndpointSubset,
) []corev1.EndpointSubset {
	var selectedKey *addressKey

	for i := range current {
		subset := current[i]

		if selectedKey == nil {
			for k := range subset.Addresses {
				addr := subset.Addresses[k]
				key := getAddressKey(&addr)

				if _, ok := readySubset[key]; ok {
					selectedKey = &key
					break
				}
			}
		}
	}
	//subset中选择一个可访问地址
	for i := range subsets {
		subset := &subsets[i]
		// move unselected addresses to not ready addresses
		naddrs := subset.Addresses
		subset.Addresses = nil

		for i := range naddrs {
			addr := naddrs[i]
			key := getAddressKey(&addr)

			if selectedKey == nil {
				selectedKey = &key
			}

			if key.ip != selectedKey.ip || key.uid != selectedKey.uid {
				subset.NotReadyAddresses = append(subset.NotReadyAddresses, addr)
			} else {
				subset.Addresses = append(subset.Addresses, addr)
			}
		}
	}

	subsets = endpointsutil.RepackSubsets(subsets)

	return subsets
}

func shouldPodBeInEndpoints(pod *corev1.Pod) bool {
	switch pod.Spec.RestartPolicy {
	case corev1.RestartPolicyNever:
		return pod.Status.Phase != corev1.PodFailed && pod.Status.Phase != corev1.PodSucceeded
	case corev1.RestartPolicyOnFailure:
		return pod.Status.Phase != corev1.PodSucceeded
	default:
		return true
	}
}
