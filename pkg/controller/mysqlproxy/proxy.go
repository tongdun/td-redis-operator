package mysqlproxy

import (
	"crypto/sha256"
	"fmt"
	"time"

	"redis-priv-operator/pkg/apis/tdb/v1alpha1"
	podutil "redis-priv-operator/third_party/kubernetes/pkg/api/v1/pod"
	"redis-priv-operator/third_party/kubernetes/pkg/util/hash"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"context"
)

func (c *Controller) syncMysqlProxy(key string) error {
	startTime := time.Now()

	defer func() {
		klog.V(4).Infof("Finished syncing mysql proxy %q. (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	p, err := c.mysqlProxyLister.MysqlProxies(ns).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		return nil
	}

	np := p.DeepCopy()

	cm, err := c.cmLister.ConfigMaps(ns).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		ncm, err := c.createConfigMap(np)
		if err != nil {
			return err
		}
		cm = ncm
	} else {
		ncm, err := c.tryUpdateConfigMap(cm, np)
		if err != nil {
			return err
		}
		cm = ncm
	}

	hasher := sha256.New()
	hash.DeepHashObject(hasher, cm.Data)
	configHash := rand.SafeEncodeString(fmt.Sprint(hasher.Sum(nil)))
	np.Status.ConfigHash = configHash

	sts, err := c.stsLister.StatefulSets(ns).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		nsts, err := c.createStatefulSet(np)
		if err != nil {
			return err
		}

		sts = nsts
	} else {
		nsts, err := c.tryUpdateStatefulSet(sts, np)
		if err != nil {
			return err
		}

		sts = nsts
	}

	svc, err := c.svcLister.Services(ns).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		nsvc, err := c.createService(np)
		if err != nil {
			return err
		}

		svc = nsvc
	} else {
		nsvc, err := c.tryUpdateService(svc, np)
		if err != nil {
			return err
		}
		svc = nsvc
	}

	np.Status.ClusterIP = svc.Spec.ClusterIP

	phase, err := c.getMysqlProxyPhase(np, sts)
	if err != nil {
		return err
	}

	np.Status.Phase = phase
	if _, err := c.extClient.TdbV1alpha1().MysqlProxies(ns).UpdateStatus(context.TODO(),np,metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *Controller) getMysqlProxyPhase(mp *v1alpha1.MysqlProxy, sts *appsv1.StatefulSet) (string, error) {
	phase := v1alpha1.MysqlProxyNotReady

	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error converting pod selector to selector: %v", err))
		return phase, nil
	}

	pods, err := c.podLister.Pods(sts.Namespace).List(selector)
	if err != nil {
		return "", err
	}

	phase = v1alpha1.MysqlProxyReady

	for i := range pods {
		pod := pods[i]

		anno, ok := pod.Annotations[v1alpha1.ConfigHashAnnotation]
		if !ok || mp.Status.ConfigHash != anno || !podutil.IsPodReady(pod) {
			phase = v1alpha1.MysqlProxyNotReady

			return phase, nil
		}
	}

	if _, err := c.epLister.Endpoints(mp.Namespace).Get(mp.Name); err != nil {
		if errors.IsNotFound(err) {
			if mp.Spec.Suspended {
				phase = v1alpha1.MysqlProxySuspended
			} else {
				phase = v1alpha1.MysqlProxyNotReady
			}

			return phase, nil
		}

		return "", err
	}

	return phase, nil
}
