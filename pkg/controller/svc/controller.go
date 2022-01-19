// Package svc defines controller for singleton service
package svc

import (
	"fmt"
	"td-redis-operator/pkg/apis/cache/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	cacheinformers "td-redis-operator/pkg/client/informers/cache/v1alpha1"
	cachelisters "td-redis-operator/pkg/client/listers/cache/v1alpha1"

	"td-redis-operator/pkg/controller"
	endpointutil "td-redis-operator/third_party/kubernetes/pkg/controller/util/endpoint"
)

// var endpointUpdatesBatchPeriod = 1 * time.Second

// ControllerOptions defines option of controller
type ControllerOptions struct {
	// KubeClient defines interface of raw kubernetes API
	KubeClient kubernetes.Interface

	// PodInformer defines informer of pod
	PodInformer coreinformers.PodInformer

	// ServiceInformer defines informer of service
	ServiceInformer coreinformers.ServiceInformer

	// EndpointsInformer defines informer of endpoints
	EndpointsInformer coreinformers.EndpointsInformer

	// RedisClusterInfromer defines informer of redis cluster
	RedisClusterInformer cacheinformers.RedisClusterInformer
}

// Controller defines controller to implements singleton service
type Controller struct {
	kubeClient kubernetes.Interface

	informersSynced []cache.InformerSynced

	queue workqueue.RateLimitingInterface

	svcLister          corelisters.ServiceLister
	podLister          corelisters.PodLister
	endpointsLister    corelisters.EndpointsLister
	redisClusterLister cachelisters.RedisClusterLister
	eventBroadcaster   record.EventBroadcaster
	eventRecorder      record.EventRecorder

	triggerTimeTracker *endpointutil.TriggerTimeTracker

	reconcilerFactory controller.ReconcilerFactory
}

// NewController returns a controller
func NewController(opt *ControllerOptions) *Controller {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: opt.KubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "td-redis-operator"})

	c := &Controller{
		kubeClient: opt.KubeClient,
		informersSynced: []cache.InformerSynced{
			opt.PodInformer.Informer().HasSynced,
			opt.ServiceInformer.Informer().HasSynced,
			opt.EndpointsInformer.Informer().HasSynced,
			opt.RedisClusterInformer.Informer().HasSynced,
		},
		queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "svc"),
		svcLister:          opt.ServiceInformer.Lister(),
		podLister:          opt.PodInformer.Lister(),
		redisClusterLister: opt.RedisClusterInformer.Lister(),
		endpointsLister:    opt.EndpointsInformer.Lister(),
		eventBroadcaster:   broadcaster,
		eventRecorder:      recorder,
		triggerTimeTracker: endpointutil.NewTriggerTimeTracker(),
		reconcilerFactory:  controller.RateLimitingReconciler,
	}

	opt.PodInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addPod,
		UpdateFunc: c.updatePod,
		DeleteFunc: c.deletePod,
	})

	opt.ServiceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addService,
		UpdateFunc: c.updateService,
		DeleteFunc: c.deleteService,
	})

	opt.RedisClusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addRedisCluster,
		UpdateFunc: c.updateRedisCluster,
		DeleteFunc: c.deleteRedisCluster,
	})

	return c
}

func (c *Controller) addPod(obj interface{}) {
	pod := obj.(*corev1.Pod)

	svcs, err := c.getPodServiceMemberships(pod)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to get pod %s/%s's service memberships: %v", pod.Namespace, pod.Name, err))
		return
	}

	for key := range svcs {
		c.queue.Add(key)
	}
}

func (c *Controller) updatePod(old, cur interface{}) {
	svcs := c.getServicesToUpdateOnPodChange(old, cur, endpointChanged)
	klog.V(5).Infof("svcs: %v", svcs)
	for key := range svcs {
		c.queue.Add(key)
	}
}
func (c *Controller) deletePod(obj interface{}) {
	pod := endpointutil.GetPodFromDeleteAction(obj)
	if pod != nil {
		c.addPod(pod)
	}
}

func (c *Controller) addService(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	_, ok := obj.(*corev1.Service).Annotations[v1alpha1.RedisClusterServiceSelectorAnnotation]
	if !ok {
		return
	}
	c.queue.Add(key)
}

func (c *Controller) updateService(old, cur interface{}) {
	c.addService(cur)
}

func (c *Controller) deleteService(obj interface{}) {
	c.addService(obj)
}

func (c *Controller) addRedisCluster(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	c.queue.Add(key)
}

func (c *Controller) updateRedisCluster(old, cur interface{}) {
	c.addRedisCluster(cur)
}

func (c *Controller) deleteRedisCluster(obj interface{}) {
	c.addRedisCluster(obj)
}

func (c *Controller) getPodServiceMemberships(pod *corev1.Pod) (sets.String, error) {
	set := sets.String{}

	svcs, err := c.svcLister.Services(pod.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for i := range svcs {
		svc := svcs[i]
		anno, ok := svc.Annotations[v1alpha1.RedisClusterServiceSelectorAnnotation]
		if !ok {
			continue
		}
		selector, err := parseSelector(anno)
		if err != nil {
			klog.Warningf("can't parse selector for service %v: %v", svc.Name, err)
			continue
		}
		if selector.Matches(labels.Set(pod.Labels)) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(svc)
			if err != nil {
				return nil, err
			}

			set.Insert(key)
		}
	}

	return set, nil
}

func (c *Controller) getServicesToUpdateOnPodChange(old, cur interface{}, endpointChanged endpointutil.EndpointsMatch) sets.String {
	newPod := cur.(*corev1.Pod)
	oldPod := old.(*corev1.Pod)
	if newPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs
		return sets.String{}
	}

	podChanged, labelsChanged := endpointutil.PodChanged(oldPod, newPod, endpointChanged)

	// If both the pod and labels are unchanged, no update is needed
	if !podChanged && !labelsChanged {
		return sets.String{}
	}

	klog.V(5).Infof("Pod is changed to sync service")

	services, err := c.getPodServiceMemberships(newPod)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Unable to get pod %s/%s's service memberships: %v", newPod.Namespace, newPod.Name, err))
		return sets.String{}
	}

	if labelsChanged {
		oldServices, err := c.getPodServiceMemberships(oldPod)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("Unable to get pod %s/%s's service memberships: %v", newPod.Namespace, newPod.Name, err))
		}
		services = endpointutil.DetermineNeededServiceUpdates(oldServices, services, podChanged)
	}

	return services
}

// Run will start the controller
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting svc controller")
	defer klog.Infof("Shutting down svc controller")

	if !cache.WaitForCacheSync(stopCh, c.informersSynced...) {
		utilruntime.HandleError(fmt.Errorf("unable to sync caches for svc controller"))
		return
	}

	klog.Infof("Cache has been synced")

	for i := 0; i < workers; i++ {
		controller.WaitUntil("svc", c.reconcilerFactory(c.queue, c.syncService), stopCh)
	}

	klog.Infof("svc controller is working")
	<-stopCh
}
