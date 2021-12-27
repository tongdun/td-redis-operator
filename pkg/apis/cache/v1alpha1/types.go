package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisClusterList defines list of redis cluster
type RedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items defines an array of redis cluster
	Items []RedisCluster `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisStandaloneList defines list of redis cluster
type RedisStandaloneList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items defines an array of redis standalone
	Items []RedisStandalone `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisClusterList defines list of redis cluster
type RedisStandbyList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items defines an array of redis masterslave
	Items []RedisStandby `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// RedisCluster defines application redis cluster
type RedisCluster struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired props of redis cluster
	// +optional
	Spec RedisClusterSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status defines the current status of redis cluster
	// +optional
	Status RedisClusterStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// RedisStandalone defines application redis standalone
type RedisStandalone struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired props of turing jupyter notebook
	// +optional
	Spec RedisStandaloneSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status defines the current status of turing jupyter notebook
	// +optional
	Status RedisStandaloneStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// RedisStandby defines application redis masterslave
type RedisStandby struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired props of redis master slave
	// +optional
	Spec RedisStandbySpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status defines the current status of redis master slave
	// +optional
	Status RedisStandbyStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// RedisClusterSpec defines spec of redis cluster
type RedisClusterSpec struct {
	// Redis memory capacity
	Capacity int `json:"capacity,omitempty" protobuf:"bytes,1,rep,name=capacity"`

	// Image defines image of redis cluster
	Image string `json:"image" protobuf:"bytes,2,name=image"`

	// Secret defines secret for redis
	Secret string `json:"secret,omitempty" protobuf:"bytes,3,opt,name=secret"`

	DC string `json:"dc,omitempty" protobuf:"bytes,4,opt,name=dc"`

	// +kubebuilder:validation:Enum="production";"staging";"demo";
	ENV string `json:"env,omitempty" protobuf:"bytes,5,opt,name=env"`

	Size int `json:"size" protobuf:"bytes,6,name=size"`

	App string `json:"app" protobuf:"bytes,7,name=app"`

	NetMode string `json:"netmode" protobuf:"bytes,8,name=netmode"`

	Vip string `json:"vip" protobuf:"bytes,9,name=vip"`

	ProxyImage string `json:"proxyimage" protobuf:"bytes,10,name=proxyimage"`

	ProxySecret string `json:"proxysecret" protobuf:"bytes,11,name=proxysecret"`

	StorageClass string `json:"storageclass" protobuf:"bytes,12,name=storageclass"`

	Realname string `json:"realname" protobuf:"bytes,13,name=realname"`

	MonitorImage string `json:"monitorimage" protobuf:"bytes,14,name=monitorimage"`
}

// RedisStandaloneSpec defines spec of redis standalone
type RedisStandaloneSpec struct {
	// Redis memory capacity
	Capacity int `json:"capacity,omitempty" protobuf:"bytes,1,rep,name=capacity"`

	// Image defines image of redis standalone
	Image string `json:"image" protobuf:"bytes,2,name=image"`

	// Secret defines secret for redis
	Secret string `json:"secret,omitempty" protobuf:"bytes,3,opt,name=secret"`

	DC string `json:"dc,omitempty" protobuf:"bytes,4,opt,name=dc"`

	ENV string `json:"env,omitempty" protobuf:"bytes,5,opt,name=env"`

	App string `json:"app" protobuf:"bytes,6,name=app"`
}

// RedisStandbySpec defines spec of redis master slave
type RedisStandbySpec struct {
	// Redis memory capacity
	Capacity int `json:"capacity,omitempty" protobuf:"bytes,1,rep,name=capacity"`

	// Image defines image of redis master slave
	Image string `json:"image" protobuf:"bytes,2,name=image"`

	// Secret defines secret for redis
	Secret string `json:"secret,omitempty" protobuf:"bytes,3,opt,name=secret"`

	DC string `json:"dc,omitempty" protobuf:"bytes,4,opt,name=dc"`

	// +kubebuilder:validation:Enum="production";"staging";"demo";
	ENV string `json:"env,omitempty" protobuf:"bytes,5,opt,name=env"`

	// SentinelImage defines image of sentinel
	SentinelImage string `json:"sentinelimage" protobuf:"bytes,6,name=sentinelimage"`

	App string `json:"app" protobuf:"bytes,7,name=app"`

	NetMode string `json:"netmode" protobuf:"bytes,8,name=netmode"`

	Vip string `json:"vip" protobuf:"bytes,9,name=vip"`

	StorageClass string `json:"storageclass" protobuf:"bytes,10,name=storageclass"`

	Realname string `json:"realname" protobuf:"bytes,11,name=realname"`

	MonitorImage string `json:"monitorimage" protobuf:"bytes,12,name=monitorimage"`
}

// Redis defines an instance of redis
type Redis struct {
	// Name defines name of redis
	Name string `json:"name" protobuf:"bytes,1,name=name"`
	// IP defines ip of redis instance
	IP string `json:"ip" protobuf:"bytes,2,name=ip"`
	// Port defines port of redis instance
	Port string `json:"port" protobuf:"bytes,3,name=port"`
}

// RedisClusterStatus defines status of redis cluster
type RedisClusterStatus struct {
	// Phase defines a phase
	Phase string `json:"phase" protobuf:"bytes,1,name=phase"`

	// ClusterIP defines internal cluster ip used by redis cluster
	ClusterIP string `json:"clusterIP" protobuf:"bytes,2,name=clusterIP"`

	// Resource create time
	GmtCreate string `json:"gmtCreate" protobuf:"bytes,3,name=gmtCreate"`

	//slots info
	Slots map[string][]string `json:"slots" protobuf:"bytes,4,name=slots"`

	Capacity int `json:"capacity,omitempty" protobuf:"bytes,5,rep,name=capacity"`

	Size int `json:"size" protobuf:"bytes,6,name=size"`

	ExternalIp string `json:"externalip" protobuf:"bytes,7,name=externalip"`
}

// RedisStandaloneStatus defines status of redis standalone
type RedisStandaloneStatus struct {
	// Phase defines a phase
	Phase string `json:"phase" protobuf:"bytes,1,name=phase"`

	// ClusterIP defines internal cluster ip used by redis standalone
	ClusterIP string `json:"clusterIP" protobuf:"bytes,2,name=clusterIP"`

	// Resource create time
	GmtCreate string `json:"gmtCreate" protobuf:"bytes,3,name=gmtCreate"`
}

// RedisStandbyStatus defines status of redis master slave
type RedisStandbyStatus struct {
	// Phase defines a phase
	Phase string `json:"phase" protobuf:"bytes,1,name=phase"`

	// ClusterIP defines internal cluster ip used by redis master
	ClusterIP string `json:"clusterIP" protobuf:"bytes,2,name=clusterIP"`

	// Resource create time
	GmtCreate string `json:"gmtCreate" protobuf:"bytes,3,name=gmtCreate"`

	Capacity int `json:"capacity,omitempty" protobuf:"bytes,4,rep,name=capacity"`

	ExternalIp string `json:"externalip" protobuf:"bytes,5,name=externalip"`
}

const (
	// RedisClusterReady defines ready phase of redis cluster
	// Ready means pods are all ready and cluster is not suspended
	RedisClusterReady = "Ready"

	// RedisClusterSuspended defiens redis cluster which is suspended
	// Suspended means pods are all ready but cluster is suspended
	RedisClusterSuspended = "Suspended"

	// RedisClusterNotReady defines not ready phase of redis cluster
	// if pods are not all ready, phase will be NotReady
	RedisClusterNotReady = "NotReady"

	// RedisClusterNodeNotReady defines pod not read in ready phasee of redis cluster
	RedisClusterNodeNotReady = "NodeNotReady"

	// RedisClusterMigrating defines cluster need scale
	RedisClusterMigrating = "Migrating"

	RedisStandbyReady = "Ready"

	RedisStandbyNotReady = "NotReady"

	RedisUpdateQuota = "UpdateQuota"
)

const (
	// SingletonServiceSelectorAnnotation defines annotation presents selector of singleton service
	// Signleton service will only select at most one ready pod ip as its endpoint
	// e.g. service.alpha.tongdun.net/singleton-selector: app=xx,name=yy
	SingletonServiceSelectorAnnotation = "service.alpha.tongdun.net/singleton-selector"

	// SingletonServiceStatusAnnotation defines annotation to disable whole svc
	SingletonServiceStatusAnnotation = "service.alpha.tongdun.net/status"

	// SingletonServiceDisabled defines value of SingletonServiceStatusAnnotation
	SingletonServiceDisabled = "disabled"

	RedisClusterServiceSelectorAnnotation = "service.alpha.tongdun.net/redis-cluster-selector"
)

const (
	// ConfigHashAnnotation defines key of config hash
	// It will be set in statefulset annotation
	ConfigHashAnnotation = "RedisCluster.alpha.tongdun.net/config-hash"
)
