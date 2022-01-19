package cluster

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"os/exec"
	"strconv"
	"strings"
	"td-redis-operator/pkg/apis/cache/v1alpha1"
	"time"
)

var (
	statefulSetGVK  = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
	redisClusterGVK = v1alpha1.SchemeGroupVersion.WithKind("RedisCluster")
)

// ConfigMapData used to render configmap
type ConfigMapData struct {
	v1alpha1.RedisCluster

	Extra ConfigMapExtraData
}

// ConfigMapExtraData defines data used by template but not in API
type ConfigMapExtraData struct {
	Secret string
}

type Predixy struct {
	Name        string
	Namespace   string
	Seed        string
	Secret      string
	ProxySecret string
	Image       string
	NetMode     string
}

func (c *Controller) createStatefulSet(p *v1alpha1.RedisCluster, real_p *v1alpha1.RedisCluster) (*appsv1.StatefulSet, error) {
	sts := appsv1.StatefulSet{}
	if err := c.statefulSetTemp.Execute(p, &sts); err != nil {
		klog.Errorf("can't render statefulset template for %v", p.Name)
		return nil, err
	}

	sts.OwnerReferences = append(sts.OwnerReferences, *metav1.NewControllerRef(real_p, redisClusterGVK))

	return c.kubeClient.AppsV1().StatefulSets(sts.Namespace).Create(context.TODO(), &sts, metav1.CreateOptions{})
}

func (c *Controller) createPredixyDeployment(p *Predixy, real_p *v1alpha1.RedisCluster) (*appsv1.Deployment, error) {
	deploy := appsv1.Deployment{}
	if err := c.predixyTemp.Execute(p, &deploy); err != nil {
		klog.Errorf("can't render predixy deployment template for %v", p.Name)
		return nil, err
	}
	deploy.OwnerReferences = append(deploy.OwnerReferences, *metav1.NewControllerRef(real_p, redisClusterGVK))
	return c.kubeClient.AppsV1().Deployments(real_p.Namespace).Create(context.TODO(), &deploy, metav1.CreateOptions{})
}

func (c *Controller) tryUpdatePredixyDeployment(deploy *appsv1.Deployment, p *v1alpha1.RedisCluster, pr *Predixy) (*appsv1.Deployment, error) {
	deployNew := appsv1.Deployment{}
	if err := c.predixyTemp.Execute(pr, &deployNew); err != nil {
		klog.Errorf("can't render deployment template for %v", p.Name)
		return nil, err
	}

	deployNew.OwnerReferences = append(deployNew.OwnerReferences, *metav1.NewControllerRef(p, redisClusterGVK))
	if apiequality.Semantic.DeepEqual(deploy.Labels, deployNew.Labels) &&
		apiequality.Semantic.DeepEqual(deploy.Annotations, deployNew.Annotations) &&
		apiequality.Semantic.DeepEqual(deploy.OwnerReferences, deployNew.OwnerReferences) {
		return deploy, nil
	}
	ndeploy := deploy.DeepCopy()
	ndeploy.Labels = deployNew.Labels
	ndeploy.Annotations = deployNew.Annotations
	ndeploy.OwnerReferences = deployNew.OwnerReferences

	if configHashChanged(deploy.Annotations, deployNew.Annotations) {
		ndeploy.Spec = deployNew.Spec
	}

	return c.kubeClient.AppsV1().Deployments(deploy.Namespace).Update(context.TODO(), ndeploy, metav1.UpdateOptions{})
}

func (c *Controller) createService(p *v1alpha1.RedisCluster) (*corev1.Service, error) {
	svc := corev1.Service{}
	if err := c.serviceTemp.Execute(p, &svc); err != nil {
		klog.Errorf("can't render service template for %v", p.Name)
		return nil, err
	}

	svc.OwnerReferences = append(svc.OwnerReferences, *metav1.NewControllerRef(p, redisClusterGVK))

	return c.kubeClient.CoreV1().Services(svc.Namespace).Create(context.TODO(), &svc, metav1.CreateOptions{})
}

func (c *Controller) createPredixyService(p *v1alpha1.RedisCluster, pr *Predixy) (*corev1.Service, error) {
	svc := corev1.Service{}
	if err := c.predixyServiceTemp.Execute(pr, &svc); err != nil {
		klog.Errorf("can't render predixy service template for %v", p.Name)
		return nil, err
	}

	svc.OwnerReferences = append(svc.OwnerReferences, *metav1.NewControllerRef(p, redisClusterGVK))

	return c.kubeClient.CoreV1().Services(svc.Namespace).Create(context.TODO(), &svc, metav1.CreateOptions{})
}

func (c *Controller) createConfigMap(p *v1alpha1.RedisCluster) (*corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{}
	d := ConfigMapData{
		RedisCluster: *p,
		/*Extra: ConfigMapExtraData{
			Secret: c.redisSecret,
		},*/
	}
	if err := c.configMapTemp.Execute(&d, &cm); err != nil {
		klog.Errorf("can't render configmap template for %v", p.Name)
		return nil, err
	}

	cm.OwnerReferences = append(cm.OwnerReferences, *metav1.NewControllerRef(p, redisClusterGVK))

	return c.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})
}

func (c *Controller) createPredixyConfigMap(p *Predixy, np *v1alpha1.RedisCluster) (*corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{}
	if err := c.predixyConfigMapTemp.Execute(&p, &cm); err != nil {
		klog.Errorf("can't render predixy configmap template for %v", p.Name)
		return nil, err
	}

	cm.OwnerReferences = append(cm.OwnerReferences, *metav1.NewControllerRef(np, redisClusterGVK))

	return c.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})
}

func (c *Controller) tryUpdateStatefulSet(sts *appsv1.StatefulSet, p *v1alpha1.RedisCluster, real_p *v1alpha1.RedisCluster) (*appsv1.StatefulSet, error) {
	stsNew := appsv1.StatefulSet{}
	if err := c.statefulSetTemp.Execute(p, &stsNew); err != nil {
		klog.Errorf("can't render statefulset template for %v", p.Name)
		return nil, err
	}

	stsNew.OwnerReferences = append(stsNew.OwnerReferences, *metav1.NewControllerRef(real_p, redisClusterGVK))
	if apiequality.Semantic.DeepEqual(sts.Labels, stsNew.Labels) &&
		apiequality.Semantic.DeepEqual(sts.Annotations, stsNew.Annotations) &&
		apiequality.Semantic.DeepEqual(sts.OwnerReferences, stsNew.OwnerReferences) {
		return sts, nil
	}

	klog.Infof("Update statefulset of redis master slave %v", p.Name)

	nsts := sts.DeepCopy()
	nsts.Labels = stsNew.Labels
	nsts.Annotations = stsNew.Annotations
	nsts.OwnerReferences = stsNew.OwnerReferences

	if configHashChanged(sts.Annotations, stsNew.Annotations) {
		nsts.Spec = stsNew.Spec
	}

	return c.kubeClient.AppsV1().StatefulSets(sts.Namespace).Update(context.TODO(), nsts, metav1.UpdateOptions{})
}

func (c *Controller) tryUpdateService(svc *corev1.Service, p *v1alpha1.RedisCluster) (*corev1.Service, error) {
	svcNew := corev1.Service{}
	if err := c.serviceTemp.Execute(p, &svcNew); err != nil {
		klog.Errorf("can't render service template for %v", p.Name)
		return nil, err
	}

	svcNew.OwnerReferences = append(svcNew.OwnerReferences, *metav1.NewControllerRef(p, redisClusterGVK))
	if apiequality.Semantic.DeepEqual(svc.Labels, svcNew.Labels) &&
		apiequality.Semantic.DeepEqual(svc.Annotations, svcNew.Annotations) &&
		apiequality.Semantic.DeepEqual(svc.OwnerReferences, svcNew.OwnerReferences) {
		return svc, nil
	}

	klog.Infof("Update service of redis cluster %v", p.Name)

	nsvc := svc.DeepCopy()
	nsvc.Labels = svcNew.Labels
	nsvc.Annotations = svcNew.Annotations
	nsvc.OwnerReferences = svcNew.OwnerReferences

	return c.kubeClient.CoreV1().Services(svc.Namespace).Update(context.TODO(), nsvc, metav1.UpdateOptions{})
}

func (c *Controller) tryUpdatePredixyService(svc *corev1.Service, p *v1alpha1.RedisCluster, pr *Predixy) (*corev1.Service, error) {
	svcNew := corev1.Service{}
	if err := c.predixyServiceTemp.Execute(pr, &svcNew); err != nil {
		klog.Errorf("can't render predixy service template for %v", p.Name)
		return nil, err
	}

	svcNew.OwnerReferences = append(svcNew.OwnerReferences, *metav1.NewControllerRef(p, redisClusterGVK))
	if apiequality.Semantic.DeepEqual(svc.Labels, svcNew.Labels) &&
		apiequality.Semantic.DeepEqual(svc.Annotations, svcNew.Annotations) &&
		apiequality.Semantic.DeepEqual(svc.OwnerReferences, svcNew.OwnerReferences) &&
		apiequality.Semantic.DeepEqual(svc.Spec.Type, svcNew.Spec.Type) {
		return svc, nil
	}

	klog.Infof("Update service of predixy %v", p.Name)

	nsvc := svc.DeepCopy()
	nsvc.Labels = svcNew.Labels
	nsvc.Annotations = svcNew.Annotations
	nsvc.OwnerReferences = svcNew.OwnerReferences
	nsvc.Spec.Type = svcNew.Spec.Type
	if nsvc.Spec.Type != "NodePort" {
		nsvc.Spec.Ports[0].NodePort = 0
	}
	return c.kubeClient.CoreV1().Services(svc.Namespace).Update(context.TODO(), nsvc, metav1.UpdateOptions{})
}

func (c *Controller) tryUpdateConfigMap(cm *corev1.ConfigMap, p *v1alpha1.RedisCluster) (*corev1.ConfigMap, error) {
	cmNew := corev1.ConfigMap{}
	d := ConfigMapData{
		RedisCluster: *p,
	}
	if err := c.configMapTemp.Execute(&d, &cmNew); err != nil {
		klog.Errorf("can't render configmap template for %v", p.Name)
		return nil, err
	}

	cmNew.OwnerReferences = append(cmNew.OwnerReferences, *metav1.NewControllerRef(p, redisClusterGVK))
	if apiequality.Semantic.DeepEqual(cm.Data, cmNew.Data) &&
		apiequality.Semantic.DeepEqual(cm.Labels, cmNew.Labels) &&
		apiequality.Semantic.DeepEqual(cm.Annotations, cmNew.Annotations) &&
		apiequality.Semantic.DeepEqual(cm.OwnerReferences, cmNew.OwnerReferences) {
		return cm, nil
	}

	klog.Infof("Update configmap of redis master slave %v", p.Name)

	ncm := cm.DeepCopy()
	ncm.Data = cmNew.Data
	ncm.Labels = cmNew.Labels
	ncm.Annotations = cmNew.Annotations
	ncm.OwnerReferences = cmNew.OwnerReferences

	return c.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Update(context.TODO(), ncm, metav1.UpdateOptions{})
}

func (c *Controller) tryUpdatePredixyConfigMap(cm *corev1.ConfigMap, p *v1alpha1.RedisCluster, predixy *Predixy) (*corev1.ConfigMap, error) {
	cmNew := corev1.ConfigMap{}
	if err := c.predixyConfigMapTemp.Execute(&predixy, &cmNew); err != nil {
		klog.Errorf("can't render configmap template for %v", p.Name)
		return nil, err
	}

	cmNew.OwnerReferences = append(cmNew.OwnerReferences, *metav1.NewControllerRef(p, redisClusterGVK))
	if apiequality.Semantic.DeepEqual(cm.Data, cmNew.Data) &&
		apiequality.Semantic.DeepEqual(cm.Labels, cmNew.Labels) &&
		apiequality.Semantic.DeepEqual(cm.Annotations, cmNew.Annotations) &&
		apiequality.Semantic.DeepEqual(cm.OwnerReferences, cmNew.OwnerReferences) {
		return cm, nil
	}

	klog.Infof("Update configmap of predixy %v", p.Name)

	ncm := cm.DeepCopy()
	ncm.Data = cmNew.Data
	ncm.Labels = cmNew.Labels
	ncm.Annotations = cmNew.Annotations
	ncm.OwnerReferences = cmNew.OwnerReferences

	return c.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Update(context.TODO(), ncm, metav1.UpdateOptions{})
}

func (c *Controller) getRedisClusterFromPod(pod *corev1.Pod) *v1alpha1.RedisCluster {
	ref := metav1.GetControllerOf(pod)
	if ref == nil {
		// No controller owns this Pod.
		return nil
	}

	if ref.Kind != statefulSetGVK.Kind {
		// Not a pod owned by a stateful set.
		return nil
	}

	sts, err := c.stsLister.StatefulSets(pod.Namespace).Get(ref.Name)
	if err != nil || sts.UID != ref.UID {
		klog.V(4).Infof("Cannot get statefulset %q for pod %q: %v", ref.Name, pod.Name, err)
		return nil
	}

	// Now find the Deployment that owns that ReplicaSet.
	ref = metav1.GetControllerOf(sts)
	if ref == nil {
		return nil
	}

	if ref.Kind != redisClusterGVK.Kind {
		return nil
	}

	mp, err := c.redisClusterLister.RedisClusters(pod.Namespace).Get(ref.Name)
	if err != nil || mp.UID != ref.UID {
		klog.V(4).Infof("Cannot get redis master slave %q for pod %q: %v", ref.Name, pod.Name, err)
		return nil
	}

	return mp
}

func (c *Controller) getRedisClusterFromEndpoints(ep *corev1.Endpoints) *v1alpha1.RedisCluster {
	mp, err := c.redisClusterLister.RedisClusters(ep.Namespace).Get(ep.Name)
	if err != nil {
		klog.V(4).Infof("Cannot get redis master slave %q: %v", ep.Name, err)
		return nil
	}
	return mp
}

func configHashChanged(a, b map[string]string) bool {
	aAnno, aOK := a[v1alpha1.ConfigHashAnnotation]
	bAnno, bOK := b[v1alpha1.ConfigHashAnnotation]

	if !aOK && !bOK {
		return false
	}

	if aOK && bOK && aAnno == bAnno {
		return false
	}

	return true
}

func (c *Controller) createRedisCluster(pods []*corev1.Pod, mp *v1alpha1.RedisCluster) error {
	podsNotinCluster := []*corev1.Pod{}
	podsInCluster := []*corev1.Pod{}
	for _, pod := range pods {
		rdb := redis.NewClient(&redis.Options{
			Addr:     pod.Status.PodIP + ":6379",
			Password: mp.Spec.Secret,
		})
		if rs, err := rdb.ClusterSlots().Result(); err != nil {
			return err
		} else {
			if len(rs) == 0 {
				podsNotinCluster = append(podsNotinCluster, pod)
			} else {
				podsInCluster = append(podsInCluster, pod)
			}
		}
	}
	if len(podsNotinCluster) == 0 {
		if mp.Spec.Capacity != mp.Status.Capacity {
			klog.Infof("start scale up %s capacity %d => %d", mp.Name, mp.Status.Capacity, mp.Spec.Capacity)

			for _, pod := range podsInCluster {
				p := redis.NewClient(&redis.Options{
					Addr:     pod.Status.PodIP + ":6379",
					Password: mp.Spec.Secret,
				})
				if _, err := p.ConfigSet("maxmemory", fmt.Sprintf("%dmb", mp.Spec.Capacity/mp.Spec.Size)).Result(); err != nil {
					klog.Errorf("%s reset maxmemory %d failed", pod.Name, mp.Spec.Capacity/mp.Spec.Size)
				}
				if _, err := p.ConfigRewrite().Result(); err != nil {
					klog.Warningf("%s save config failed", pod.Name)
				}
				p.Close()
			}
			klog.Infof("%s scale up finished", mp.Name)
			mp.Status.Capacity = mp.Spec.Capacity
		}
		return nil
	}
	gps := map[string][]string{}
	gps_masters := []string{}
	gps_ips := map[string]string{}
	gps_master_slave := map[string][]string{}
	for _, pod := range pods {
		gps[pod.Labels["RESOURCE_ID"]] = append(gps[pod.Labels["RESOURCE_ID"]], pod.Status.PodIP+":6379")
		gps_ips[pod.Status.PodIP+":6379"] = pod.Labels["RESOURCE_ID"]
	}
	for c, ps := range gps {
		gps_masters = append(gps_masters, ps[0])
		gps_master_slave[ps[0]] = append(gps_master_slave[ps[0]], ps[1], c)
	}
	if len(mp.Status.Slots) != 0 {
		klog.Infof("initialized cluster，slot %v", mp.Status.Slots)
		gp_empty := map[string][]*corev1.Pod{}
		for _, pod := range podsNotinCluster {
			rdb := redis.NewClient(&redis.Options{
				Addr:     pod.Status.PodIP + ":6379",
				Password: mp.Spec.Secret,
			})
			if mp.Status.Slots[pod.Labels["RESOURCE_ID"]] == nil {
				gp_empty[pod.Labels["RESOURCE_ID"]] = append(gp_empty[pod.Labels["RESOURCE_ID"]], pod)
			}
			for _, slotrange := range mp.Status.Slots[pod.Labels["RESOURCE_ID"]] {
				start, _ := strconv.Atoi(strings.Split(slotrange, "-")[0])
				end, _ := strconv.Atoi(strings.Split(slotrange, "-")[1])
				if _, err := rdb.ClusterAddSlotsRange(start, end).Result(); err != nil {
					klog.Infof("assign slot range %s to %s failed %v", slotrange, pod.Name, err)
					if _, err2 := rdb.ClusterResetHard().Result(); err2 != nil {
						return err
					}
					klog.Infof("reset %s finished", pod.Name)
					return err
				}
			}
			time.Sleep(5 * time.Second)
			for _, p := range podsInCluster {
				if p.Labels["RESOURCE_ID"] == pod.Labels["RESOURCE_ID"] {
					continue
				}
				if _, err := rdb.ClusterMeet(p.Status.PodIP, "6379").Result(); err != nil {
					klog.Warningf("%s meet %s failed %v", pod.Name, p.Name, err)
				}
			}
			//DELETE INVALID NODE
			forgetBadNode(pods, gps_ips, mp.Spec.Secret)
		}
		//NEED SCALE?
		if mp.Spec.Size > mp.Status.Size {
			mp.Status.Phase = v1alpha1.RedisUpdateQuota
			if rp, err := c.extClient.CacheV1alpha1().RedisClusters(mp.Namespace).UpdateStatus(context.TODO(), mp, metav1.UpdateOptions{}); err != nil {
				return err
			} else {
				*mp = *rp
			}
			m := &redis.Client{}
			master := ""
			klog.Infof("%s recognized need scale cluster", mp.Name)
			for gp, pods := range gp_empty {
				if len(pods) != 2 {
					klog.Warningf("%s group ready nodes not reach 2", gp)
					break
				}
				master = pods[0].Status.PodIP + ":6379"
				slave := pods[1].Status.PodIP + ":6379"
				if err := clusterCheck(mp.Spec.Secret, master); err != nil {
					return err
				}
				klog.Infof("start add %s group", gp)
				m = redis.NewClient(&redis.Options{
					Addr:     master,
					Password: mp.Spec.Secret,
				})
				node_id, err := m.Do("cluster", "myid").String()
				if err != nil {
					return err
				}
				s := redis.NewClient(&redis.Options{
					Addr:     slave,
					Password: mp.Spec.Secret,
				})
				if _, err := s.ClusterReplicate(node_id).Result(); err != nil {
					klog.Infof("slave %s sync master %s failed %v", slave, master, err)
					return err
				}
				klog.Infof("slave %s sync master %s success", slave, master)
				time.Sleep(2 * time.Second)
			}
			//START SCALE
			if err := clusterCheck(mp.Spec.Secret, master); err != nil {
				return err
			}
			cmdstr := fmt.Sprintf("redis-cli -a %s --cluster rebalance --cluster-use-empty-masters %s", mp.Spec.Secret, master)
			klog.Infof(cmdstr)
			cmd := exec.Command("bash", "-c", cmdstr)
			if err := cmd.Start(); err != nil {
				klog.Errorf("%s size %s to %s rebalance failed %v", mp.Name, mp.Status.Size, mp.Spec.Size, err)
			}
			if err := cmd.Wait(); err != nil {
				klog.Error(err)
				return err
			}
			mp.Status.Capacity = mp.Spec.Capacity
			mp.Status.Size = mp.Spec.Size
			mp.Status.Slots = map[string][]string{}
			if rs, err := m.ClusterSlots().Result(); err != nil {
				klog.Warningf("get slot info failed %v", err)
				return err
			} else {
				for _, slot := range rs {
					mp.Status.Slots[gps_ips[slot.Nodes[0].Addr]] = append(mp.Status.Slots[gps_ips[slot.Nodes[0].Addr]], fmt.Sprintf("%d-%d", slot.Start, slot.End))
				}
			}
			klog.Infof("slot info updated %v", mp.Status.Slots)
			//reset maxmemory
			for _, pod := range podsInCluster {
				p := redis.NewClient(&redis.Options{
					Addr:     pod.Status.PodIP + ":6379",
					Password: mp.Spec.Secret,
				})
				if _, err := p.ConfigSet("maxmemory", fmt.Sprintf("%dmb", mp.Spec.Capacity/mp.Status.Size)).Result(); err != nil {
					klog.Errorf("%s reset maxmemory %d failed", pod.Name, mp.Spec.Capacity/mp.Status.Size)
				}
				if _, err := p.ConfigRewrite().Result(); err != nil {
					klog.Warningf("%s save config failed", pod.Name)
				}
				p.Close()
			}

		}
		mp.Status.Phase = v1alpha1.RedisClusterReady
		return nil
	}
	//operator image should have redis-client
	klog.Infof("initialize cluster %v", gps_masters)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo yes|redis-cli -a %s --cluster create %s", mp.Spec.Secret, strings.Join(gps_masters, " ")))
	if err := cmd.Run(); err != nil {
		return err
	}
	klog.Infof("initialize cluster success，add node replica %v", gps_master_slave)
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    gps_masters,
		Password: mp.Spec.Secret,
	})
	mp.Status.Slots = map[string][]string{}
	if rs, err := rdb.ClusterSlots().Result(); err != nil {
		return err
	} else {
		for _, slot := range rs {
			mp.Status.Slots[gps_master_slave[slot.Nodes[0].Addr][1]] = append(mp.Status.Slots[gps_master_slave[slot.Nodes[0].Addr][1]], fmt.Sprintf("%d-%d", slot.Start, slot.End))
		}
	}
	for _, master := range gps_masters {
		m := redis.NewClient(&redis.Options{
			Addr:     master,
			Password: mp.Spec.Secret,
		})
		klog.Infof("create master connection %s", master)
		node_id, err := m.Do("cluster", "myid").String()
		if err != nil {
			return err
		}
		klog.Infof("get master %s node id %s", master, node_id)
		s := redis.NewClient(&redis.Options{
			Addr:     gps_master_slave[master][0],
			Password: mp.Spec.Secret,
		})
		mip := strings.Split(master, ":")[0]
		mport := strings.Split(master, ":")[1]
		klog.Infof("slave %s meet master %s", gps_master_slave[master][0], master)
		if _, err := s.ClusterMeet(mip, mport).Result(); err != nil {
			klog.Warningf("initialized phase %s meet %s failed", gps_master_slave[master][0], master)
			return err
		}
		time.Sleep(2 * time.Second)
		klog.Infof("slave %s sync master %s", gps_master_slave[master][0], master)
		if _, err := s.ClusterReplicate(node_id).Result(); err != nil {
			klog.Infof("slave %s sync master %s failed %v", gps_master_slave[master][0], master, err)
			return err
		}
		klog.Infof("slave %s sync master %s success", gps_master_slave[master][0], master)

	}
	klog.Infof("add node replica success")
	mp.Status.Size = mp.Spec.Size
	mp.Status.Capacity = mp.Spec.Capacity
	mp.Status.Phase = v1alpha1.RedisClusterReady
	mp.Status.GmtCreate = time.Now().Format("2006-01-02 15:04:05")
	return nil
}

func forgetBadNode(pods []*corev1.Pod, gps_ips map[string]string, password string) {
	for _, pod := range pods {
		p := redis.NewClient(&redis.Options{
			Addr:     pod.Status.PodIP + ":6379",
			Password: password,
		})
		if rs, err := p.ClusterNodes().Result(); err != nil {
			p.Close()
			continue
		} else {
			for _, info := range strings.Split(rs, "\n") {
				if len(strings.Split(info, " ")) < 2 {
					continue
				}
				node_id := strings.Split(info, " ")[0]
				ipport := strings.Split(strings.Split(info, " ")[1], "@")[0]
				if gps_ips[ipport] == "" {
					klog.Infof("delete invalid node %s", ipport)
					for _, pod2 := range pods {
						p2 := redis.NewClient(&redis.Options{
							Addr:     pod2.Status.PodIP + ":6379",
							Password: password,
						})
						p2.ClusterForget(node_id)
						p2.Close()
					}
				}
			}
		}
	}
}

func clusterCheck(auth string, ins string) error {
	err_count := 0
	for {
		time.Sleep(2 * time.Second)
		if err_count > 30 {
			return errors.New("cluster check too much times...")
		}
		cmdstr := fmt.Sprintf("redis-cli -a %s --cluster check %s", auth, ins)
		klog.Info(cmdstr)
		cmd := exec.Command("bash", "-c", cmdstr)
		if err := cmd.Run(); err != nil {
			klog.Warningf("%v,Performing Cluster Check Continue...", err)
			err_count = err_count + 1
			continue
		}
		break
	}
	return nil
}
