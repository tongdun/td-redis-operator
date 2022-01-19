package masterslave

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/go-redis/redis"
	"k8s.io/apimachinery/pkg/labels"
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

func (c *Controller) syncRedisStandby(key string) error {
	startTime := time.Now()

	defer func() {
		klog.V(4).Infof("Finished syncing redis master slave %q. (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	p, err := c.redisStandbyLister.RedisStandbies(ns).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		klog.Infof("redis standby %s was deleted", name)
		return nil
	}
	senti_name := "sentinel-" + p.Spec.App
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
	//create sentinel sts
	sts, err := c.stsLister.StatefulSets(ns).Get(senti_name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		nsts, err := c.createSentiStatefulSet(np)
		if err != nil {
			return err
		}

		sts = nsts
	} else {
		nsts, err := c.tryUpdateSentiStatefulSet(sts, np)
		if err != nil {
			return err
		}

		sts = nsts
	}
	svc, err := c.svcLister.Services(ns).Get(senti_name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		nsvc, err := c.createSentiService(np)
		if err != nil {
			return err
		}

		svc = nsvc
	} else {
		nsvc, err := c.tryUpdateSentiService(svc, np)
		if err != nil {
			return err
		}
		svc = nsvc
	}

	//create redis sts
	sts, err = c.stsLister.StatefulSets(ns).Get(name)
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

	svc, err = c.svcLister.Services(ns).Get(name)
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

	np.Status.ClusterIP = svc.Spec.ClusterIP + ":6379"
	if np.Spec.NetMode == "NodePort" {
		np.Status.ExternalIp = fmt.Sprintf("%s:%d", np.Spec.Vip, svc.Spec.Ports[0].NodePort)
	} else {
		np.Status.ExternalIp = ""
	}
	/*phase, err := c.getRedisMsPhase(np, sts)
	if err != nil {
		return err
	}*/
	if np.Status.Capacity != np.Spec.Capacity && np.Status.Capacity != 0 {
		np.Status.Phase = v1alpha1.RedisUpdateQuota
		if np, err = c.extClient.CacheV1alpha1().RedisStandbies(ns).UpdateStatus(context.TODO(), np, metav1.UpdateOptions{}); err != nil {
			return err
		}
		key := map[string]string{"CLUSTER": np.Name}
		selector := labels.SelectorFromSet(labels.Set(key))
		if pods, err := c.podLister.Pods(sts.Namespace).List(selector); err == nil {
			for _, pod := range pods {
				p := redis.NewClient(&redis.Options{
					Addr:     pod.Status.PodIP + ":6379",
					Password: np.Spec.Secret,
				})
				if _, err := p.ConfigSet("maxmemory", fmt.Sprintf("%dmb", np.Spec.Capacity)).Result(); err != nil {
					klog.Errorf("%s reset maxmemory %dmb failed", pod.Name, np.Spec.Capacity)
					return err
				}
				if _, err := p.ConfigRewrite().Result(); err != nil {
					klog.Warningf("%s save config failed", pod.Name)
					return err
				}
				p.Close()

			}
		} else {
			return err
		}
		np.Status.Phase = v1alpha1.RedisStandbyReady
	}
	switch np.Status.Phase {
	case "":
		time.Sleep(10 * time.Second)
		p := redis.NewClient(&redis.Options{
			Addr:     svc.Spec.ClusterIP + ":6379",
			Password: np.Spec.Secret,
		})
		if _, err := p.Ping().Result(); err != nil {
			klog.Warningf("%s service not ready", np.Name)
			return err
		}
		np.Status.Phase = v1alpha1.RedisStandbyReady
		np.Status.Capacity = np.Spec.Capacity
		if np.Status.GmtCreate == "" {
			np.Status.GmtCreate = time.Now().Format("2006-01-02 15:04:05")
		}
		klog.Infof("%s ready", np.Name)
		break
	case v1alpha1.RedisStandbyReady:
		np.Status.Capacity = np.Spec.Capacity
		break
	case v1alpha1.RedisUpdateQuota:
		np.Status.Phase = v1alpha1.RedisStandbyReady
		break
	}
	if _, err := c.extClient.CacheV1alpha1().RedisStandbies(ns).UpdateStatus(context.TODO(), np, metav1.UpdateOptions{}); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}

func (c *Controller) getRedisStandbyPhase(mp *v1alpha1.RedisStandby, sts *appsv1.StatefulSet) (string, error) {
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
