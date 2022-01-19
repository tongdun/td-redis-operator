// Package config defines operator config struct
package config

import (
	"k8s.io/client-go/informers"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coordinformers "k8s.io/client-go/informers/coordination/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"td-redis-operator/pkg/client/clientset"
	extinformers "td-redis-operator/pkg/client/informers"
	redisinformers "td-redis-operator/pkg/client/informers/cache/v1alpha1"
	"td-redis-operator/pkg/template"
)

// Config defines operator config
type Config struct {
	// KubeClient defines kubernetes client interface
	KubeClient kubernetes.Interface

	// ExtClient defines extension client interface
	ExtClient clientset.Interface

	// KubeInformerFactory defines informer factory
	KubeInformerFactory informers.SharedInformerFactory

	// LeaseKubeInformerFactory defines informer factory to watch kube-node-lease
	LeaseKubeInformerFactory informers.SharedInformerFactory

	// ExtInformerFactory defines extension informer factory
	ExtInformerFactory extinformers.SharedInformerFactory

	// PodInformer defines pod informer
	PodInformer coreinformers.PodInformer

	// ServiceInformer defines service informer
	ServiceInformer coreinformers.ServiceInformer

	// EndpointsInformer defines endpoints informer
	EndpointsInformer coreinformers.EndpointsInformer

	// ConfigMapInformer defines configmap informer
	ConfigMapInformer coreinformers.ConfigMapInformer

	// StatefulSetInformer defines statefulset informer
	StatefulSetInformer appsinformers.StatefulSetInformer

	DeploymentInformer appsinformers.DeploymentInformer

	// LeaseInformer defines informer for lease
	LeaseInformer coordinformers.LeaseInformer

	// RedisStandaloneInformer defines redis standalone informer
	RedisStandaloneInformer redisinformers.RedisStandaloneInformer

	// RedisMasterslaveInfomer defines redis master slave informer
	RedisStandbyInformer redisinformers.RedisStandbyInformer

	// RedisClusterInformer defines redis cluster informer
	RedisClusterInformer redisinformers.RedisClusterInformer

	// StatefulSetTemplate defines template of mysql proxy
	StatefulSetTemplate *template.Template

	// RedisStandalone StatefulSetTemplate defines template of Redis standalone
	RedisStandaloneStatefulSetTemplate *template.Template

	// RedisStandby statefulsetTemplate defines template of Redis master slave
	RedisStandbyStatefulSetTemplate *template.Template

	SentinelStatefulSetTemplate *template.Template

	// RedisCluster statefulsetTemplate defines template of Redis cluster
	RedisClusterStatefulSetTemplate *template.Template

	// ServiceTemplate defines template of mysql proxy service
	ServiceTemplate *template.Template

	// RedisStandaloneServiceTemplate defines template of Redis standalone
	RedisStandaloneServiceTemplate *template.Template

	// RedisStandbyServiceTemplate defines template of Redis master slave
	RedisStandbyServiceTemplate *template.Template

	SentiServiceTemplate *template.Template
	// RedisClusterServiceTemplate defines template of Redis cluster
	RedisClusterServiceTemplate *template.Template

	ClusterPredixyConfigMapTemplate *template.Template

	PredixyTemplate *template.Template

	PredixyServiceTemplate *template.Template

	// ConfigMapTemplate defines template of mysql proxy configmap
	ConfigMapTemplate *template.Template

	RedisStandaloneConfigMapTemplate *template.Template
	RedisStandbyConfigMapTemplate    *template.Template
	RedisClusterConfigMapTemplate    *template.Template

	// MysqlSecret defines secret for mysql
	MysqlSecret string

	// RedisSecret defines secret for redis
	RedisSecret string

	// EnableHostEndpoint defines whether enable host endpoint controller
	EnableHostEndpoint bool
}
