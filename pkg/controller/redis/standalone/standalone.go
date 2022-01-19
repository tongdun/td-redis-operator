package standalone

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"td-redis-operator/pkg/apis/cache/v1alpha1"
	"td-redis-operator/third_party/kubernetes/pkg/util/hash"
)

func (c *Controller) syncRedisStandalone(key string) error {
	startTime := time.Now()

	defer func() {
		klog.V(4).Infof("Finished syncing redis standalone %q. (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	p, err := c.redisStandaloneLister.RedisStandalones(ns).Get(name)
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

	/*phase, err := c.getRedisStandalonePhase(np, sts)
	if err != nil {
		return err
	}*/

	np.Status.Phase = "运行中"
	if _, err := c.extClient.CacheV1alpha1().RedisStandalones(ns).UpdateStatus(context.TODO(), np, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *Controller) getRedisStandalonePhase(mp *v1alpha1.RedisStandalone, sts *appsv1.StatefulSet) (string, error) {
	phase := v1alpha1.RedisClusterNotReady

	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error converting pod selector to selector: %v", err))
		return phase, nil
	}

	_, err = c.podLister.Pods(sts.Namespace).List(selector)
	if err != nil {
		return "", err
	}

	phase = v1alpha1.RedisClusterReady

	/*if _, err := c.epLister.Endpoints(mp.Namespace).Get(mp.Name); err != nil {
		if errors.IsNotFound(err) {
			if mp.Spec.Suspended {
				phase = v1alpha1.RedisClusterSuspended
			} else {
				phase = v1alpha1.RedisClusterNotReady
			}

			return phase, nil
		}

		return "", err
	}*/

	return phase, nil
}
