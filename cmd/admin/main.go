package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"td-redis-operator/pkg/client/clientset"
	. "td-redis-operator/pkg/conf"
	"td-redis-operator/pkg/middlewares/common"
	"td-redis-operator/pkg/middlewares/luc"
	"td-redis-operator/pkg/template"
	"time"
)

func main() {
	restConfig, err := clientcmd.BuildConfigFromFlags("", Cfg.Kubeconfig)
	if err != nil {
		panic(fmt.Errorf("can't parse kubeconfig from (%v)", Cfg.Kubeconfig))
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(fmt.Errorf("can't new kubernetes client: %v", err))
	}

	extClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		panic(fmt.Errorf("can't new extension client: %v", err))
	}

	c := Client{
		KubeClient:        kubeClient,
		ExtClient:         extClient,
		Namespace:         Cfg.Namespace,
		RedisStandbyImage: Cfg.Standbyimage,
		SentiImage:        Cfg.Sentiimage,
		RedisClusterImage: Cfg.Clusterimage,
		ProxyImage:        Cfg.Proxyimage,
		RedisSecret:       Cfg.Redissecret,
		StorageClass:      Cfg.Storageclass,
		Vip:               Cfg.Vip,
		MonitorImage:      Cfg.MonitorImage,
	}
	r := gin.Default()
	for _, mw := range Cfg.Middlewares {
		switch mw {
		case "luc":
			luc.SSO = Cfg.Luc
			r.Use(luc.Luc)
		default:
			r.Use(common.Common)
		}
	}
	r.GET("/api/v1alpha2/redis", c.getRedis)
	r.GET("/api/v1alpha2/redisall", c.getRedisAll)
	r.PUT("/api/v1alpha2/changeowner", c.changeOwner)
	r.GET("/api/v1alpha2/redis/slowlog/:name", c.getSlowlog)
	r.GET("/api/v1alpha2/redis/operlog/:name", c.getOperLog)
	r.GET("/api/v1alpha2/redis/config/:name", c.getRedisConf)
	r.DELETE("/api/v1alpha2/redis", c.deleteRedis)
	r.POST("/api/v1alpha2/redis", c.createRedis)
	r.PUT("/api/v1alpha2/redis", c.updateRedis)
	r.PUT("/api/v1alpha2/redis/flush", c.flushRedis)
	r.PUT("/api/v1alpha2/redis/config/:name", c.updateRedisConf)
	nodes_resource = map[string]NodeResource{}
	used_memory = map[string]string{}
	go func() {
		for {
			GetUsedMemory(&c)
			GetReqLimitByNode(&c.KubeClient)
			time.Sleep(5 * time.Minute)
		}
	}()
	r.Run(":8088")
}

// Client defines client visit kubernetes
type Client struct {
	KubeClient        kubernetes.Interface
	ExtClient         clientset.Interface
	Namespace         string
	RedisStandbyImage string
	SentiImage        string
	RedisClusterImage string
	ProxyImage        string
	RedisSecret       string
	Vip               string
	StorageClass      string
	MonitorImage      string
	Template          *template.Template
}

// JSONError defines a json error
type JSONError struct {
	// Status will always be false
	Status  bool   `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// JSONMarshalError defines an error string of marshal error
var JSONMarshalError = []byte(`{"status":false,"reason":"JSONMarshalError","message":"can't parse error to json"}`)

// WriteJSON writes json into response
func WriteJSON(w http.ResponseWriter, code int, obj interface{}) {
	body, err := json.Marshal(obj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteWithLog(w, JSONMarshalError)

		return
	}

	w.WriteHeader(code)
	WriteWithLog(w, body)
}

// WriteWithLog logs if write error
func WriteWithLog(w io.Writer, body []byte) {
	_, err := w.Write(body)
	if err != nil {
		fmt.Printf("write error: %v\n", err)
	}
}
