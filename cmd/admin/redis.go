package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
	"td-redis-operator/pkg/apis/cache/v1alpha1"
	"td-redis-operator/pkg/conf"
	"td-redis-operator/pkg/dbhelper"
	"td-redis-operator/pkg/logger"
	. "td-redis-operator/pkg/redis"
	. "td-redis-operator/pkg/web"
	"time"
)

type SlowLog struct {
	Ts   string `json:"ts"`
	Cost int64  `json:"cost"`
	Cmd  string `json:"cmd"`
	Src  string `json:"src"`
}

const (
	RedisCluster = "cluster"
	RedisStandby = "standby"
	Op_Update    = "update"
	Op_Create    = "create"
)

func (c *Client) deleteRedis(co *gin.Context) {
	var req Redis
	realname := co.GetString("realname")
	isdba := co.GetBool("dba")
	if err := co.ShouldBindJSON(&req); err != nil {
		co.JSON(http.StatusBadRequest, Result(err.Error(), false))
		return
	}
	if req.Name == "" {
		co.JSON(http.StatusBadRequest, Result("invalid resource name", false))
		return
	}
	logid := logger.LogOper(realname, req.Name, "delete")
	switch req.Kind {
	case RedisStandby:
		name := c.Redis2Standby(&req).Name
		if rs, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("lookup %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		} else {
			if rs.Spec.Realname != realname && !isdba {
				co.JSON(http.StatusBadRequest, Result("no privileges", false))
				logger.UpdateOperStatus(logid, logger.OperFailed)
				return
			}

		}
		if err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("delete %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
	case RedisCluster:
		name := c.Redis2Cluster(&req).Name
		if rs, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("delete %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		} else {
			if rs.Spec.Realname != realname && !isdba {
				co.JSON(http.StatusBadRequest, Result("no privileges", false))
				logger.UpdateOperStatus(logid, logger.OperFailed)
				return
			}
		}
		if err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Delete(context.TODO(), c.Redis2Cluster(&req).Name, metav1.DeleteOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("delete %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
	default:
		co.JSON(http.StatusBadRequest, Result("invalied resouce name", false))
		logger.UpdateOperStatus(logid, logger.OperFailed)
		return
	}
	delete(used_memory, req.Name)
	logger.UpdateOperStatus(logid, logger.OperSuccess)
	co.JSON(http.StatusOK, Result("ok", true))
}

func (c *Client) flushRedis(co *gin.Context) {
	var req Redis
	realname := co.GetString("realname")
	if err := co.ShouldBindJSON(&req); err != nil {
		co.JSON(http.StatusBadRequest, Result(err.Error(), false))
		return
	}
	logid := logger.LogOper(realname, req.Name, "empty clean")
	switch req.Kind {
	case RedisStandby:
		c := redis.NewClient(&redis.Options{
			Addr:     req.ClusterIP,
			Password: req.Secret,
		})
		if _, err := c.FlushAllAsync().Result(); err != nil {
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
	case RedisCluster:
		key := map[string]string{"APP": req.Name}
		pods, err := c.getPodsWithlabel(key)
		if err != nil {
			logger.ERROR(err.Error())
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		for _, pod := range pods {
			c := redis.NewClient(&redis.Options{
				Addr:     pod.Status.PodIP + ":6379",
				Password: c.RedisSecret,
			})
			if _, err := c.FlushAllAsync().Result(); err != nil {
				co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
				logger.UpdateOperStatus(logid, logger.OperFailed)
				return
			}
			c.Close()

		}

	}
	logger.UpdateOperStatus(logid, logger.OperSuccess)
	co.JSON(http.StatusOK, Result("ok", true))
}

func (c *Client) getSlowlog(co *gin.Context) {
	var slowlogs []SlowLog
	name := co.Param("name")
	namelist := strings.Split(name, "-")
	if name == "" || len(namelist) < 2 {
		co.JSON(http.StatusInternalServerError, Result("invalid", false))
		return
	}
	instance_type := strings.Split(name, "-")[0]
	switch instance_type {
	case RedisStandby:
		r, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Get(context.TODO(), "redis-"+name, metav1.GetOptions{})
		if err != nil {
			logger.ERROR(err.Error())
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			return
		}
		c := redis.NewClient(&redis.Options{
			Addr:     r.Status.ClusterIP,
			Password: r.Spec.Secret,
		})
		rs, err := c.Do("slowlog", "get", "100").Result()
		if err != nil {
			logger.ERROR(err.Error())
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			return
		}
		slowlogs = getSlowlogList(rs)
	case RedisCluster:
		key := map[string]string{"APP": name}
		pods, err := c.getPodsWithlabel(key)
		rss := []interface{}{}
		if err != nil {
			logger.ERROR(err.Error())
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			return
		}
		for _, pod := range pods {
			c := redis.NewClient(&redis.Options{
				Addr:     pod.Status.PodIP + ":6379",
				Password: c.RedisSecret,
			})
			rs, err := c.Do("slowlog", "get", "10").Result()
			if err != nil {
				logger.ERROR(err.Error())
				co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
				return
			}
			if len(rs.([]interface{})) == 0 {
				continue
			}
			rss = append(rss, rs.([]interface{})...)
		}
		slowlogs = getSlowlogList(rss)
	}
	co.JSON(http.StatusOK, slowlogs)
}

func (c *Client) getOperLog(co *gin.Context) {
	name := co.Param("name")
	namelist := strings.Split(name, "-")
	if name == "" || len(namelist) < 2 {
		co.JSON(http.StatusInternalServerError, Result("invalid", false))
		return
	}
	operlogs, err := logger.GetOperLogs(name)
	if err != nil {
		logger.ERROR(err.Error())
		return
	}
	co.JSON(http.StatusOK, operlogs)
}

func getSlowlogList(rs interface{}) []SlowLog {
	slowlogs := []SlowLog{}
	for _, log := range rs.(([]interface{})) {
		slowlog := SlowLog{}
		infos := log.([]interface{})
		slowlog.Ts = time.Unix(infos[1].(int64), 0).Format("2006-01-02 15:04:05")
		slowlog.Cost = infos[2].(int64)
		cmdlist := []string{}
		for _, i := range infos[3].([]interface{}) {
			cmdlist = append(cmdlist, i.(string))
		}
		slowlog.Cmd = strings.Join(cmdlist, " ")
		slowlog.Src = infos[4].(string)
		slowlogs = append(slowlogs, slowlog)
	}
	return slowlogs
}

func (c *Client) getRedis(co *gin.Context) {
	rediss := []Redis{}
	RedisStandies, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
		return
	}
	for _, r := range RedisStandies.Items {
		if r.Spec.Realname != co.GetString("realname") && !co.GetBool("dba") {
			continue
		}
		rediss = append(rediss, *c.Standby2Redis(&r))
	}
	RedisClusters, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
		return
	}
	for _, r := range RedisClusters.Items {
		if r.Spec.Realname != co.GetString("realname") && !co.GetBool("dba") {
			continue
		}
		rediss = append(rediss, *c.Cluster2Redis(&r))
	}
	co.JSON(http.StatusOK, rediss)
}

func (c *Client) getRedisAll(co *gin.Context) {
	rediss := []Redis{}
	rediss_final := []Redis{}
	dc := strings.ToLower(co.Query("dc"))
	env := strings.ToLower(co.Query("env"))
	RedisStandies, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
		return
	}
	for _, r := range RedisStandies.Items {
		rediss = append(rediss, *c.Standby2Redis(&r))
	}
	RedisClusters, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
		return
	}
	for _, r := range RedisClusters.Items {
		rediss = append(rediss, *c.Cluster2Redis(&r))
	}
	if redissv1, err := dbhelper.GetRedisv1(); err != nil {
		logger.WARN(err.Error())
	} else {
		rediss = append(rediss, redissv1...)
	}
	if redissmv, err := dbhelper.GetRedisVM(); err != nil {
		logger.WARN(err.Error())
	} else {
		rediss = append(rediss, redissmv...)
	}
	for _, redis := range rediss {
		if dc != "" && dc != strings.ToLower(redis.Dc) {
			continue
		}
		if env != "" && env != strings.ToLower(redis.Env) {
			continue
		}
		rediss_final = append(rediss_final, redis)
	}
	co.JSON(http.StatusOK, rediss_final)

}

func (c *Client) changeOwner(co *gin.Context) {
	name := co.Query("name")
	newowner := co.Query("newowner")
	cloud_name := "redis-" + name
	if name == "" || newowner == "" {
		co.JSON(http.StatusInternalServerError, Result("invalid request", false))
		return
	}
	if r, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Get(context.TODO(), cloud_name, metav1.GetOptions{}); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			return
		}
	} else {
		//更新CR
		r.Spec.Realname = newowner
		if _, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Update(context.TODO(), r, metav1.UpdateOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			return

		}
		co.JSON(http.StatusOK, Result("update success，cachetype:redis standby", true))
		return
	}
	if r, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Get(context.TODO(), cloud_name, metav1.GetOptions{}); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			return
		}
	} else {
		r.Spec.Realname = newowner
		if _, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Update(context.TODO(), r, metav1.UpdateOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			return
		}
		co.JSON(http.StatusOK, Result("update success，cachetype:redis cluster", true))
		return
	}
	//更新v1 redis
	if err := dbhelper.ChangeOwnerV1(name, newowner); err != nil {
		if err.Error() != "Not found" {
			co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
			return
		}
	} else {
		co.JSON(http.StatusOK, Result("update success，cachetype:redisv1", true))
		return
	}
	//查找是否是虚拟机资源
	vms, err := dbhelper.GetRedisVM()
	if err != nil {
		co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
		return
	}
	for _, vm := range vms {
		if vm.Name == name {
			if err := dbhelper.FixRedisVMOwner(name, newowner); err != nil {
				co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
				return
			}
			co.JSON(http.StatusOK, Result("update success，cachetype:redisvm", true))
			return
		}
	}
	//返回为空
	co.JSON(http.StatusOK, Result("unknow resource", false))
	return

}

func (c *Client) createRedis(co *gin.Context) {
	var req Redis
	realname := co.GetString("realname")
	req.Realname = realname
	if err := co.ShouldBindJSON(&req); err != nil {
		co.JSON(http.StatusBadRequest, Result(err.Error(), false))
		return
	}
	req.Realname = realname
	switch req.Kind {
	case RedisStandby:
		if err := roughResource(RedisStandby, int64(req.Capacity)); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("create standby %s failed:%v", req.Name, err), false))
			return
		}
		if _, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Create(context.TODO(), c.Redis2Standby(&req), metav1.CreateOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("create standby %s failed:%v", req.Name, err), false))
			return
		}
	case RedisCluster:
		if err := roughResource(RedisCluster, int64(req.Capacity)); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("create cluster %s failed:%v", req.Name, err), false))
			return
		}
		if _, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Create(context.TODO(), c.Redis2Cluster(&req), metav1.CreateOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("create cluster %s failed:%v", req.Name, err), false))
			return
		}
	default:
		co.JSON(http.StatusBadRequest, Result("invalid resource", false))
		return
	}
	co.JSON(http.StatusOK, Result("ok", true))
}

func (c *Client) updateRedis(co *gin.Context) {
	var req Redis
	realname := co.GetString("realname")
	isdba := co.GetBool("dba")
	if err := co.ShouldBindJSON(&req); err != nil {
		co.JSON(http.StatusBadRequest, Result(err.Error(), false))
		return
	}
	logid := logger.LogOper(realname, req.Name, "update quota")
	if req.Phase == v1alpha1.RedisUpdateQuota {
		co.JSON(http.StatusInternalServerError, Result(req.Name+"updating quota,please waiting", false))
		logger.UpdateOperStatus(logid, logger.OperFailed)
		return
	}
	switch req.Kind {
	case RedisStandby:
		name := c.Redis2Standby(&req).Name
		if rs, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("look for %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		} else {
			if rs.Spec.Realname != realname && !isdba {
				co.JSON(http.StatusBadRequest, Result("no privileges", false))
				logger.UpdateOperStatus(logid, logger.OperFailed)
				return
			}

		}
		if err := roughResource(RedisStandby, int64(req.Capacity)); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("update standby %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		if _, err := c.allowUpdate(&req); err != nil {
			co.JSON(http.StatusBadRequest, Result(err.Error(), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		rs, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("get standby %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		rs.Spec.Capacity = req.Capacity
		rs.Spec.NetMode = req.NetMode
		rs.Spec.Vip = c.Vip
		if _, err := c.ExtClient.CacheV1alpha1().RedisStandbies(c.Namespace).Update(context.TODO(), rs, metav1.UpdateOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("update standby %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		break
	case RedisCluster:
		name := c.Redis2Cluster(&req).Name
		if rs, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("look for %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		} else {
			if rs.Spec.Realname != realname && !isdba {
				co.JSON(http.StatusBadRequest, Result("no privileges", false))
				logger.UpdateOperStatus(logid, logger.OperFailed)
				return
			}
		}
		if err := roughResource(RedisCluster, int64(req.Capacity)); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("update cluster %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		size, err := c.allowUpdate(&req)
		if err != nil {
			co.JSON(http.StatusBadRequest, Result(err.Error(), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		cs, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Get(context.TODO(), c.Redis2Cluster(&req).Name, metav1.GetOptions{})
		if err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("gain cluster %s failed:%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		cs.Spec.Size = size
		cs.Spec.Capacity = req.Capacity
		cs.Spec.NetMode = req.NetMode
		cs.Spec.Vip = c.Vip
		if _, err := c.ExtClient.CacheV1alpha1().RedisClusters(c.Namespace).Update(context.TODO(), cs, metav1.UpdateOptions{}); err != nil {
			co.JSON(http.StatusInternalServerError, Result(fmt.Sprintf("update cluster %s :%v", req.Name, err), false))
			logger.UpdateOperStatus(logid, logger.OperFailed)
			return
		}
		break
	default:
		co.JSON(http.StatusBadRequest, Result("invalid resource name", false))
		logger.UpdateOperStatus(logid, logger.OperFailed)
		return
	}
	logger.UpdateOperStatus(logid, logger.OperSuccess)
	co.JSON(http.StatusOK, Result("ok", true))
}

func (c *Client) getRedisConf(co *gin.Context) {
	name := co.Param("name")
	if name == "" {
		co.JSON(http.StatusBadRequest, Result("empty resource name", false))
		return
	} else {
		name = fmt.Sprintf("redis-%s", name)
	}
	redisconf, err := c.KubeClient.CoreV1().ConfigMaps(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logger.ERROR(err.Error())
		co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
		return
	}
	co.JSON(http.StatusOK, conf.GetRedisConfList(redisconf.Data["redis.conf"]))
}

func (c *Client) updateRedisConf(co *gin.Context) {
	var req []conf.RedisConfEntry
	realname := co.GetString("realname")
	if err := co.ShouldBindJSON(&req); err != nil {
		co.JSON(http.StatusBadRequest, Result(err.Error(), false))
		return
	}
	name := co.Param("name")
	if name == "" {
		co.JSON(http.StatusBadRequest, Result("empty resource name", false))
		return
	}
	redisconf, err := c.KubeClient.CoreV1().ConfigMaps(c.Namespace).Get(context.TODO(), fmt.Sprintf("redis-%s", name), metav1.GetOptions{})
	if err != nil {
		co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
		return
	}
	cmmp := conf.Getcmmap(redisconf.Data["redis.conf"])
	password := strings.Trim(cmmp["requirepass"], "\"")
	key := map[string]string{"CLUSTER": fmt.Sprintf("redis-%s", name)}
	pods, err := c.getPodsWithlabel(key)
	logid := logger.LogOper(realname, name, "修改配置文件")
	for _, pod := range pods {
		c := redis.NewClient(&redis.Options{
			Addr:     pod.Status.PodIP + ":6379",
			Password: password,
		})
		for _, entry := range req {
			v := ""
			switch entry.Value.(type) {
			case string:
				v = fmt.Sprintf("%s", entry.Value)
			case int:
				v = fmt.Sprintf("%d", int(entry.Value.(int)))
			case float32:
				v = fmt.Sprintf("%d", int(entry.Value.(float32)))
			case float64:
				v = fmt.Sprintf("%d", int(entry.Value.(float64)))
			}
			if _, err := c.ConfigSet(entry.Name, v).Result(); err != nil {
				logger.ERROR(err.Error())
				continue
			}
		}
		if _, err := c.ConfigRewrite().Result(); err != nil {
			logger.ERROR(err.Error())
			continue
		}
		c.Close()
	}
	redisconf.Data["redis.conf"] = conf.GetRedisCmstr(req, redisconf.Data["redis.conf"])
	_, err = c.KubeClient.CoreV1().ConfigMaps(c.Namespace).Update(context.TODO(), redisconf, metav1.UpdateOptions{})
	if err != nil {
		co.JSON(http.StatusInternalServerError, Result(err.Error(), false))
		logger.UpdateOperStatus(logid, logger.OperFailed)
		return
	}
	logger.UpdateOperStatus(logid, logger.OperSuccess)
	co.JSON(http.StatusOK, "")
}
