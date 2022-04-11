// Package options defines operator options
package options

import (
	"fmt"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"td-redis-operator/cmd/operator/app/config"
	"td-redis-operator/pkg/client/clientset"
	extinformers "td-redis-operator/pkg/client/informers"
	"td-redis-operator/pkg/template"
)

// Options defines running options of CI/CD for ML
type Options struct {
	Kubeconfig string

	Namespace string

	MysqlSecret string

	EnableHostEndpoint bool
}

// NewOptions returns new running options
func NewOptions() *Options {
	opt := &Options{
		Kubeconfig: "",
		// 监听的namespace
		Namespace: "redis",
	}

	return opt
}

// AddFlags adds flags for ML options
func (opt *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&opt.Kubeconfig, "kubeconfig", opt.Kubeconfig, "kubeconfig of cluster")
	fs.StringVar(&opt.Namespace, "namespace", opt.Namespace, "namespace of operator, if empty, all namespaces will be watched")
	fs.StringVar(&opt.MysqlSecret, "mysql-secret", opt.MysqlSecret, "secret for mysql")
	fs.BoolVar(&opt.EnableHostEndpoint, "enable-host-endpoint", opt.EnableHostEndpoint, "enable host endpoint, only support calico cni")
}

// Config parse options to config of operator
func (opt *Options) Config() (*config.Config, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", opt.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("can't parse kubeconfig from (%v)", opt.Kubeconfig)
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("can't new kube client: %v", err)
	}

	extClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("can't new extension client: %v", err)
	}

	var (
		kubeInformerOpts []informers.SharedInformerOption
		extInformerOpts  []extinformers.SharedInformerOption
	)

	if len(opt.Namespace) != 0 {
		kubeInformerOpts = append(kubeInformerOpts, informers.WithNamespace(opt.Namespace))
		extInformerOpts = append(extInformerOpts, extinformers.WithNamespace(opt.Namespace))
	}

	kubeInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0, kubeInformerOpts...)
	extInformerFactory := extinformers.NewSharedInformerFactoryWithOptions(extClient, 0, extInformerOpts...)

	podInformer := kubeInformerFactory.Core().V1().Pods()
	serviceInformer := kubeInformerFactory.Core().V1().Services()
	endpointsInformer := kubeInformerFactory.Core().V1().Endpoints()
	configMapInformer := kubeInformerFactory.Core().V1().ConfigMaps()
	statefulSetInformer := kubeInformerFactory.Apps().V1().StatefulSets()
	deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()
	redisStandInformer := extInformerFactory.Cache().V1alpha1().RedisStandalones()
	redisStandbyInformer := extInformerFactory.Cache().V1alpha1().RedisStandbies()
	redisClusterInformer := extInformerFactory.Cache().V1alpha1().RedisClusters()

	

	redisStandaloneStatefulsetTemp, err := template.NewTemplate("/statefulset_redis_standalone.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create redis standalone statefulset template")
	}

	redisStandbyStatefulsetTemp, err := template.NewTemplate("/statefulset_redis_standby.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create redis master slave statefulset template")
	}

	sentinelStatefulsetTemp, err := template.NewTemplate("/statefulset_sentinel.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create redis master slave statefulset template")
	}

	redisClusterStatefulsetTemp, err := template.NewTemplate("/statefulset_redis_cluster.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create redis cluster statefulset template")
	}

	serviceTemp, err := template.NewTemplate("/service.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create service template")
	}

	redisStandaloneServiceTemp, err := template.NewTemplate("/service_redis_standalone.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create service redis standalone template")
	}

	redisStandbyServiceTemp, err := template.NewTemplate("/service_redis_standby.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create service redis ms template")
	}

	sentiServiceTemp, err := template.NewTemplate("/service_sentinel.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create service redis ms template")
	}

	redisClusterServiceTemp, err := template.NewTemplate("/service_redis_cluster.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create service redis cluster template")
	}

	redisStandaloneConfigMapTemp, err := template.NewTemplate("/configmap_redis_standalone.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create redis standalone configmap template")
	}

	redisStandbyConfigMapTemp, err := template.NewTemplate("/configmap_redis_standby.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create redis standby configmap template")
	}

	redisClusterConfigMapTemp, err := template.NewTemplate("/configmap_redis_cluster.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create redis cluster configmap template")
	}

	clusterPredixyConfigMapTemp, err := template.NewTemplate("/configmap_cluster_predixy.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create redis cluster predixy configmap template")
	}
	predixyTemp, err := template.NewTemplate("/deployment_predixy.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create predixy template")
	}
	predixyServiceTemp, err := template.NewTemplate("/service_predixy.tmpl")
	if err != nil {
		return nil, fmt.Errorf("cant' create predixy service template")
	}

	leaseKubeInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0, informers.WithNamespace(corev1.NamespaceNodeLease))

	leaseInformer := leaseKubeInformerFactory.Coordination().V1().Leases()

	c := &config.Config{
		KubeClient: kubeClient,
		ExtClient:  extClient,

		KubeInformerFactory:      kubeInformerFactory,
		LeaseKubeInformerFactory: leaseKubeInformerFactory,
		ExtInformerFactory:       extInformerFactory,

		PodInformer:         podInformer,
		ServiceInformer:     serviceInformer,
		EndpointsInformer:   endpointsInformer,
		ConfigMapInformer:   configMapInformer,
		StatefulSetInformer: statefulSetInformer,
		DeploymentInformer:  deploymentInformer,
		LeaseInformer:       leaseInformer,

		RedisStandaloneInformer: redisStandInformer,
		RedisStandbyInformer:    redisStandbyInformer,
		RedisClusterInformer:    redisClusterInformer,

		StatefulSetTemplate:                statefulSetTemp,
		RedisStandaloneStatefulSetTemplate: redisStandaloneStatefulsetTemp,
		RedisStandbyStatefulSetTemplate:    redisStandbyStatefulsetTemp,
		RedisClusterStatefulSetTemplate:    redisClusterStatefulsetTemp,
		ServiceTemplate:                    serviceTemp,
		RedisStandaloneServiceTemplate:     redisStandaloneServiceTemp,
		RedisStandbyServiceTemplate:        redisStandbyServiceTemp,
		SentiServiceTemplate:               sentiServiceTemp,
		SentinelStatefulSetTemplate:        sentinelStatefulsetTemp,
		RedisClusterServiceTemplate:        redisClusterServiceTemp,
		ClusterPredixyConfigMapTemplate:    clusterPredixyConfigMapTemp,
		PredixyTemplate:                    predixyTemp,
		PredixyServiceTemplate:             predixyServiceTemp,
		RedisStandaloneConfigMapTemplate:   redisStandaloneConfigMapTemp,
		RedisStandbyConfigMapTemplate:      redisStandbyConfigMapTemp,
		RedisClusterConfigMapTemplate:      redisClusterConfigMapTemp,

		MysqlSecret: opt.MysqlSecret,

		EnableHostEndpoint: opt.EnableHostEndpoint,
	}

	return c, nil
}
