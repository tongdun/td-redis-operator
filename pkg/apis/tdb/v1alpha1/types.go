package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MysqlProxyList defines list of mysql proxy
type MysqlProxyList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items defines an array of mysql proxy
	Items []MysqlProxy `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MysqlProxy defines application mysql proxy
type MysqlProxy struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired props of turing jupyter notebook
	// +optional
	Spec MysqlProxySpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status defines the current status of turing jupyter notebook
	// +optional
	Status MysqlProxyStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// MysqlProxySpec defines spec of mysql proxy
type MysqlProxySpec struct {
	// Mysqls defines mysql instances
	// +optional
	Mysqls []Mysql `json:"mysqls,omitempty" protobuf:"bytes,1,rep,name=mysqls"`

	// Image defines image of mysql proxy
	Image string `json:"image" protobuf:"bytes,2,name=image"`

	// Suspended defines mysql proxy is suspended
	Suspended bool `json:"suspended,omitempty" protobuf:"bytes,3,opt,name=suspended"`

	// Secret defines secret for mysql
	Secret string `json:"secret,omitempty" protobuf:"bytes,4,opt,name=secret"`
}

// Mysql defines an instance of mysql
type Mysql struct {
	// Name defines name of mysql
	Name string `json:"name" protobuf:"bytes,1,name=name"`
	// IP defines ip of mysql instance
	IP string `json:"ip" protobuf:"bytes,2,name=ip"`
	// Port defines port of mysql instance
	Port string `json:"port" protobuf:"bytes,3,name=port"`
}

// MysqlProxyStatus defines status of mysql proxy
type MysqlProxyStatus struct {
	// Phase defines a phase
	Phase string `json:"phase" protobuf:"bytes,1,name=phase"`

	// ClusterIP defines internal cluster ip used by mysql proxy
	ClusterIP string `json:"clusterIP" protobuf:"bytes,2,name=clusterIP"`

	// ConfigHash defines hash of config file
	// Pods will be updated only when config hash is changed
	ConfigHash string `json:"configHash" protobuf:"bytes,3,name=configHash"`
}

// NOTE(bo.liub): use Conditions in the future
const (
	// MysqlProxyReady defines ready phase of mysql proxy
	// Ready means pods are all ready and proxy is not suspended
	MysqlProxyReady = "Ready"

	// MysqlProxySuspended defiens mysql proxy which is suspended
	// Suspended means pods are all ready but proxy is suspended
	MysqlProxySuspended = "Suspended"

	// MysqlProxyNotReady defines not ready phase of mysql proxy
	// if pods are not all ready, phase will be NotReady
	MysqlProxyNotReady = "NotReady"
)

const (
	// RedisClusterServiceSelectorAnnotation defines annotation presents selector of RedisCluster service
	// RedisCluster service will only select at most one ready pod ip as its endpoint
	// e.g. service.alpha.tongdun.net/redis-cluster-selector: app=xx,name=yy
	RedisClusterServiceSelectorAnnotation = "service.alpha.tongdun.net/redis-cluster-selector"

	// RedisClusterServiceStatusAnnotation defines annotation to disable whole svc
	RedisClusterServiceStatusAnnotation = "service.alpha.tongdun.net/status"

	// SingletonServiceDisabled defines value of SingletonServiceStatusAnnotation
	RedisClusterServiceDisabled = "disabled"
)

const (
	// ConfigHashAnnotation defines key of config hash
	// It will be set in statefulset annotation
	ConfigHashAnnotation = "mysqlproxy.alpha.tongdun.net/config-hash"
)
