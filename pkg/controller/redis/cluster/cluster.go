package cluster

import (
	"context"
	"crypto/sha256"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"td-redis-operator/pkg/apis/cache/v1alpha1"
	podutil "td-redis-operator/third_party/kubernetes/pkg/api/v1/pod"
	"td-redis-operator/third_party/kubernetes/pkg/util/hash"
	"time"
)

func (c *Controller) syncRedisCluster(key string) error {
	startTime := time.Now()

	defer func() {
		klog.V(4).Infof("Finished syncing redis cluster %q. (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	p, err := c.redisClusterLister.RedisClusters(ns).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		klog.Infof("redis cluster %s was deleted", name)
		return nil
	}
	np := p.DeepCopy()

	cm, err := c.cmLister.ConfigMaps(ns).Get(name)
	cmp := p.DeepCopy()
	cmp.Spec.Capacity = cmp.Spec.Capacity / cmp.Spec.Size
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		ncm, err := c.createConfigMap(cmp)
		if err != nil {
			return err
		}
		cm = ncm
	} else {
		ncm, err := c.tryUpdateConfigMap(cm, cmp)
		if err != nil {
			return err
		}
		cm = ncm
	}

	hasher := sha256.New()
	sts := &appsv1.StatefulSet{}
	hash.DeepHashObject(hasher, cm.Data)
	for idx := 0; idx < np.Spec.Size; idx++ {
		group_sts := np.DeepCopy()
		group_sts.Name = fmt.Sprintf("%s-%d", group_sts.Name, idx)
		sts, err = c.stsLister.StatefulSets(ns).Get(group_sts.Name)
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}

			nsts, err := c.createStatefulSet(group_sts, np)
			if err != nil {
				return err
			}

			sts = nsts
		} else {
			nsts, err := c.tryUpdateStatefulSet(sts, group_sts, np)
			if err != nil {
				return err
			}
			sts = nsts
		}
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

	if err = c.getRedisClusterPhase(np, sts); err != nil {
		return err
	}

	predixy := &Predixy{
		Name:        "predixy-" + np.Name,
		Namespace:   np.Namespace,
		Seed:        svc.Spec.ClusterIP,
		Secret:      np.Spec.Secret,
		ProxySecret: np.Spec.ProxySecret,
		NetMode:     np.Spec.NetMode,
		Image:       np.Spec.ProxyImage,
	}
	cm, err = c.cmLister.ConfigMaps(ns).Get(predixy.Name)
	cmp = p.DeepCopy()
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		ncm, err := c.createPredixyConfigMap(predixy, cmp)
		if err != nil {
			return err
		}
		cm = ncm
	} else {
		ncm, err := c.tryUpdatePredixyConfigMap(cm, cmp, predixy)
		if err != nil {
			return err
		}
		cm = ncm
	}

	predixy_dp, err := c.deployLister.Deployments(ns).Get(predixy.Name)
	dp := p.DeepCopy()
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		_, err := c.createPredixyDeployment(predixy, dp)
		if err != nil {
			return err
		}
	} else {
		_, err := c.tryUpdatePredixyDeployment(predixy_dp, dp, predixy)
		if err != nil {
			return err
		}
	}

	svc, err = c.svcLister.Services(ns).Get(predixy.Name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		nsvc, err := c.createPredixyService(np, predixy)
		if err != nil {
			return err
		}

		svc = nsvc
	} else {
		nsvc, err := c.tryUpdatePredixyService(svc, np, predixy)
		if err != nil {
			return err
		}
		svc = nsvc
	}

	np.Status.ClusterIP = svc.Spec.ClusterIP + ":6379"
	if np.Spec.NetMode == "NodePort" {
		np.Status.ExternalIp = fmt.Sprintf("%s:%d", np.Spec.Vip, svc.Spec.Ports[0].NodePort)
	} else {
		np.Status.ExternalIp = ""
	}

	if _, err := c.extClient.CacheV1alpha1().RedisClusters(ns).UpdateStatus(context.TODO(), np, metav1.UpdateOptions{}); err != nil {
		klog.Infof("%s update failed:%v", name, err)
		return err
	}
	//klog.Infof("%s更新成功", name)
	time.Sleep(1 * time.Second)
	return nil
}

func (c *Controller) getRedisClusterPhase(mp *v1alpha1.RedisCluster, sts *appsv1.StatefulSet) error {
	key := map[string]string{"APP": mp.Spec.App}
	selector := labels.SelectorFromSet(labels.Set(key))
	pods, err := c.podLister.Pods(sts.Namespace).List(selector)
	if err != nil {
		return err
	}
	if len(pods) != mp.Spec.Size*2 {
		return errors.NewBadRequest(mp.Name + " pod number not fit expect:" + fmt.Sprintf("%d", len(pods)))
	}

	for i := range pods {
		pod := pods[i]
		if !podutil.IsPodReady(pod) {
			return errors.NewBadRequest(pod.Name + " pod not ready")
		}
	}
	if err = c.createRedisCluster(pods, mp); err != nil {
		return err
	}
	return nil
}
