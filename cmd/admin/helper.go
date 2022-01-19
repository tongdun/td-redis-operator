package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"strconv"
	"strings"
	"td-redis-operator/pkg/apis/cache/v1alpha1"
	"td-redis-operator/pkg/logger"
	. "td-redis-operator/pkg/redis"
	podutil "td-redis-operator/third_party/kubernetes/pkg/api/v1/pod"
)

type NodeResource struct {
	Node           v1.Node
	Cpulimitleft   int64
	Cpurequestleft int64
	Memlimitleft   int64
	Memrequestleft int64
}

var nodes_resource map[string]NodeResource

var used_memory map[string]string

func calSize(cap int) int {
	if cap <= 3*32*1024 {
		return 3
	}
	if cap%(32*1024) != 0 {
		return cap/(32*1024) + 1
	}
	return cap / (32 * 1024)
}

func (c *Client) Redis2Standby(r *Redis) *v1alpha1.RedisStandby {
	return &v1alpha1.RedisStandby{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("redis-%s", r.Name),
			Namespace: c.Namespace,
		},
		Spec: v1alpha1.RedisStandbySpec{
			Capacity:      r.Capacity,
			Image:         c.RedisStandbyImage,
			Secret:        r.Secret,
			DC:            r.Dc,
			ENV:           r.Env,
			SentinelImage: c.SentiImage,
			App:           r.Name,
			NetMode:       r.NetMode,
			Vip:           c.Vip,
			StorageClass:  c.StorageClass,
			Realname:      r.Realname,
			MonitorImage:  c.MonitorImage,
		},
	}
}

func (c *Client) Standby2Redis(rs *v1alpha1.RedisStandby) *Redis {
	var host string
	var port int
	if info := strings.Split(rs.Status.ClusterIP, ":"); len(info) == 2 {
		host = strings.Split(rs.Status.ClusterIP, ":")[0]
		port, _ = strconv.Atoi(strings.Split(rs.Status.ClusterIP, ":")[1])
	}
	usedmemory := used_memory["redis-"+rs.Spec.App]
	if usedmemory == "" {
		usedmemory = "0"
	}
	return &Redis{
		Name:       rs.Spec.App,
		Phase:      rs.Status.Phase,
		ClusterIP:  rs.Status.ClusterIP,
		ExternalIP: rs.Status.ExternalIp,
		GmtCreate:  rs.Status.GmtCreate,
		Capacity:   rs.Status.Capacity,
		Secret:     rs.Spec.Secret,
		Dc:         rs.Spec.DC,
		Env:        rs.Spec.ENV,
		Kind:       RedisStandby,
		NetMode:    rs.Spec.NetMode,
		Realname:   rs.Spec.Realname,
		MemoryUsed: usedmemory,
		Host:       host,
		Port:       port,
	}
}

func (c *Client) Redis2Cluster(r *Redis) *v1alpha1.RedisCluster {
	return &v1alpha1.RedisCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("redis-%s", r.Name),
			Namespace: c.Namespace,
		},
		Spec: v1alpha1.RedisClusterSpec{
			Capacity:     r.Capacity,
			Image:        c.RedisClusterImage,
			ProxyImage:   c.ProxyImage,
			Secret:       c.RedisSecret,
			ProxySecret:  r.Secret,
			DC:           r.Dc,
			ENV:          r.Env,
			App:          r.Name,
			Size:         calSize(r.Capacity),
			NetMode:      r.NetMode,
			Vip:          c.Vip,
			StorageClass: c.StorageClass,
			Realname:     r.Realname,
			MonitorImage: c.MonitorImage,
		},
	}
}

func (c *Client) Cluster2Redis(cs *v1alpha1.RedisCluster) *Redis {
	var host string
	var port int
	if info := strings.Split(cs.Status.ClusterIP, ":"); len(info) == 2 {
		host = strings.Split(cs.Status.ClusterIP, ":")[0]
		port, _ = strconv.Atoi(strings.Split(cs.Status.ClusterIP, ":")[1])
	}
	usedmemory := used_memory["redis-"+cs.Spec.App]
	if usedmemory == "" {
		usedmemory = "0"
	}
	return &Redis{
		Name:       cs.Spec.App,
		Phase:      cs.Status.Phase,
		ClusterIP:  cs.Status.ClusterIP,
		ExternalIP: cs.Status.ExternalIp,
		GmtCreate:  cs.Status.GmtCreate,
		Capacity:   cs.Status.Capacity,
		Secret:     cs.Spec.ProxySecret,
		Dc:         cs.Spec.DC,
		Env:        cs.Spec.ENV,
		Kind:       RedisCluster,
		NetMode:    cs.Spec.NetMode,
		Realname:   cs.Spec.Realname,
		MemoryUsed: usedmemory,
		Host:       host,
		Port:       port,
	}
}

func (c *Client) getPodsWithlabel(key map[string]string) ([]v1.Pod, error) {
	selector := labels.SelectorFromSet(labels.Set(key)).String()
	pods, err := c.KubeClient.CoreV1().Pods(c.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func (c *Client) allowUpdate(r *Redis) (int, error) {
	if r.ClusterIP == "" {
		return 0, errors.New("Empty ClusterIP")
	}
	switch r.Kind {
	case RedisStandby:
		if memmb, err := getUsedMemMb(r.Secret, r.ClusterIP); err != nil {
			return 0, err
		} else {
			if r.Capacity <= memmb {
				return 0, errors.New(fmt.Sprintf("master slave%s forbidden update:expect %d,used %d", r.Name, r.Capacity, memmb))
			}
		}
		return 0, nil
	case RedisCluster:
		key := map[string]string{"APP": r.Name}
		var avg_mb int
		cs, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Get(context.TODO(), c.Redis2Cluster(r).Name, metav1.GetOptions{})
		if err != nil {
			return 0, err
		} else {
			avg_mb = r.Capacity / cs.Status.Size
		}
		if pods, err := c.KubeClient.CoreV1().Pods(c.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.Set(key).String()}); err != nil {
			return 0, err
		} else {
			for _, pod := range pods.Items {
				if memmb, err := getUsedMemMb(c.RedisSecret, pod.Status.PodIP+":6379"); err != nil {
					return 0, err
				} else {
					if avg_mb <= memmb {
						return 0, errors.New(fmt.Sprintf("Cluster node %s forbidden update:expect %d,used %d", pod.Name, avg_mb, memmb))
					}
				}

			}
		}
		var size int
		if calSize(r.Capacity) <= cs.Spec.Size {
			size = cs.Spec.Size
		} else {
			size = calSize(r.Capacity)
		}
		return size, nil
	default:
		return 0, errors.New("unknow resource type")
	}
	return 0, nil
}

func getUsedMemMb(secret string, ip string) (int, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     ip,
		Password: secret,
	})
	if rs, err := c.Info("memory").Result(); err != nil {
		return 0, err
	} else {
		var mem int64
		for _, s := range strings.Split(rs, "\r\n") {
			if strings.Split(s, ":")[0] == "used_memory" {
				mem, err = strconv.ParseInt(strings.Split(s, ":")[1], 10, 64)
				if err != nil {
					return 0, err
				}
				strInt64 := strconv.FormatInt(mem/1024/1024, 10)
				mem_mb, _ := strconv.Atoi(strInt64)
				return mem_mb, nil
			}
		}

	}
	return 0, errors.New("bad return")
}

func roughResource(resource_type string, memory int64) error {
	fit_nodes := 0
	ready_nodes := 0
	if resource_type == RedisStandby {
		for _, node_resource := range nodes_resource {
			if !isNodeReady(node_resource.Node) || len(node_resource.Node.Spec.Taints) != 0 {
				continue
			}
			ready_nodes = ready_nodes + 1
			if node_resource.Cpurequestleft >= 2100 && node_resource.Memrequestleft > memory*(1024*1024)*2 {
				fit_nodes = fit_nodes + 1
			}
			if fit_nodes >= 2 && ready_nodes >= 3 {
				return nil
			}
		}
		return errors.New(fmt.Sprintf("resource not enough，please choose smaller specs or add node,fit %d,want %d,ready %d", fit_nodes, 2, 3))
	}
	if resource_type == RedisCluster {
		for _, node_resource := range nodes_resource {
			if !isNodeReady(node_resource.Node) || len(node_resource.Node.Spec.Taints) != 0 {
				continue
			}
			if node_resource.Cpurequestleft >= 4200 && node_resource.Memrequestleft >= (1024*1024)*memory/int64(calSize(int(memory))) {
				if node_resource.Cpurequestleft >= 8400 && node_resource.Memrequestleft >= (1024*1024)*2*memory/int64(calSize(int(memory))) {
					fit_nodes = fit_nodes + 2
				} else {
					fit_nodes = fit_nodes + 1
				}
			}
			if fit_nodes >= 2*calSize(int(memory)) {
				return nil
			}
		}
		return errors.New(fmt.Sprintf("resource not enough，please choose smaller specs or add node,fit %d,want %d", fit_nodes, 2*calSize(int(memory))))
	}
	return errors.New("error type")
}

func isNodeReady(node v1.Node) bool {

	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			return true
		}
	}
	return false
}

func GetReqLimitByNode(kc *kubernetes.Interface) {
	nodes, err := (*kc).CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.ERROR(err.Error())
		return
	}
	temp := make(map[string]NodeResource)
	for _, node := range nodes.Items {
		fieldSelector, _ := fields.ParseSelector("spec.nodeName=" + node.Name + ",status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
		var CPUlimit, CPUrequest, Memlimit, Memreq int64
		nodeNonTerminatedPodsList, err := (*kc).CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
		if err != nil {
			klog.Error(err)
		}
		for _, pod := range nodeNonTerminatedPodsList.Items {
			CPUlimit = CPUlimit + pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()
			CPUrequest = CPUrequest + pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()
			Memlimit = Memlimit + pod.Spec.Containers[0].Resources.Limits.Memory().Value()
			Memreq = Memreq + pod.Spec.Containers[0].Resources.Requests.Memory().Value()
		}
		temp[node.Name] = NodeResource{
			Node:           node,
			Cpulimitleft:   node.Status.Allocatable.Cpu().MilliValue() - CPUlimit,
			Cpurequestleft: node.Status.Allocatable.Cpu().MilliValue() - CPUrequest,
			Memlimitleft:   node.Status.Allocatable.Memory().Value() - Memlimit,
			Memrequestleft: node.Status.Allocatable.Memory().Value() - Memreq,
		}

	}
	nodes_resource = temp

}

func GetUsedMemory(client *Client) {
	temp := make(map[string]string)
	RedisStandies, err := client.ExtClient.CacheV1alpha1().RedisStandbies(client.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.ERROR(err.Error())
		return
	}
	for _, r := range RedisStandies.Items {
		if r.Status.ClusterIP == "" {
			logger.WARN(fmt.Sprintf("%s empty clusterip", r.Name))
			continue
		}
		c := redis.NewClient(&redis.Options{
			Addr:     r.Status.ClusterIP,
			Password: r.Spec.Secret,
		})
		if rs, err := c.Info("memory").Result(); err != nil {
			logger.WARN(err.Error())
			continue
		} else {
			for _, s := range strings.Split(rs, "\r\n") {
				if strings.Split(s, ":")[0] == "used_memory_human" {
					temp[r.Name] = strings.Split(s, ":")[1]
					break
				}
			}

		}
		c.Close()
	}
	RedisClusters, err := client.ExtClient.CacheV1alpha1().RedisClusters(client.Namespace).List(context.TODO(), metav1.ListOptions{})
	for _, r := range RedisClusters.Items {
		var total_int64 int64
		cluster_ok := true
		for k, _ := range r.Status.Slots {
			pod, err := client.KubeClient.CoreV1().Pods(client.Namespace).Get(context.TODO(), k+"-0", metav1.GetOptions{})
			if !podutil.IsPodReady(pod) || err != nil {
				cluster_ok = false
				break
			}
			c := redis.NewClient(&redis.Options{
				Addr:     pod.Status.PodIP + ":6379",
				Password: r.Spec.Secret,
			})
			if rs, err := c.Info("memory").Result(); err != nil {
				logger.WARN(err.Error())
				cluster_ok = false
				break
			} else {
				for _, s := range strings.Split(rs, "\r\n") {
					if strings.Split(s, ":")[0] == "used_memory" {
						int64, err := strconv.ParseInt(strings.Split(s, ":")[1], 10, 64)
						if err != nil {
							logger.WARN(err.Error())
							cluster_ok = false
							break
						}
						total_int64 = total_int64 + int64
						break
					}
				}

			}
		}
		if cluster_ok {
			temp[r.Name] = getByteHuman(total_int64)
		} else {
			temp[r.Name] = used_memory[r.Name]
		}
	}
	used_memory = temp
}

func getByteHuman(total int64) string {
	if total/1024 < 1 {
		return fmt.Sprintf("%d", total)
	}
	if total/1024/1024 < 1 {
		return fmt.Sprintf("%dK", total/1024)
	}
	if total/1024/1024/1024 < 1 {
		return fmt.Sprintf("%dM", total/1024/1024)
	}
	if total/1024/1024/1024/1024 < 1 {
		return fmt.Sprintf("%dG", total/1024/1024/1024)
	}
	return fmt.Sprintf("%d", total)
}
