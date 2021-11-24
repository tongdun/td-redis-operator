package redis

type Redis struct {
	Name         string `json:"name"`
	Phase        string `json:"phase"`
	ClusterIP    string `json:"clusterIp"`
	ExternalIP   string `json:"externalIp"`
	GmtCreate    string `json:"gmtCreate"`
	Capacity     int    `json:"capacity"`
	MemoryUsed   string `json:"memoryused"`
	Secret       string `json:"secret"`
	Dc           string `json:"dc"`
	Env          string `json:"env"`
	Kind         string `json:"kind"`
	NetMode      string `json:"netMode"`
	StorageClass string `json:"storageclass"`
	Realname     string `json:"realname"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
}
