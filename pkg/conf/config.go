package conf

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var Cfg *Config

type Config struct {
	Logger       map[string]string
	Kubeconfig   string
	Namespace    string
	Standbyimage string
	Sentiimage   string
	Clusterimage string
	Proxyimage   string
	Redissecret  string
	Vip          string
	Storageclass string
	Middlewares  []string
	Luc          string
	MonitorImage string
	Mon          map[string]string
}

func init() {
	//初始化系统配置文件
	Cfg = &Config{}
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	var configdir string
	pflag.CommandLine.StringVar(&configdir, "configdir", configdir, "configdir of redis operator")
	pflag.Parse()
	viper.AddConfigPath(configdir)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	Cfg.Logger = viper.GetStringMapString("logger")
	Cfg.Mon = viper.GetStringMapString("mon")
	Cfg.Kubeconfig = viper.GetString("kubeconfig")
	Cfg.Namespace = viper.GetString("namespace")
	Cfg.Standbyimage = viper.GetString("standbyimage")
	Cfg.Sentiimage = viper.GetString("sentiimage")
	Cfg.Clusterimage = viper.GetString("clusterimage")
	Cfg.Proxyimage = viper.GetString("proxyimage")
	Cfg.Redissecret = viper.GetString("redissecret")
	Cfg.Vip = viper.GetString("vip")
	Cfg.Storageclass = viper.GetString("storageclass")
	Cfg.Middlewares = viper.GetStringSlice("middlewares")
	Cfg.Luc = viper.GetString("luc")
	Cfg.MonitorImage = viper.GetString("monitorimage")
	//初始化redis配置文件编辑项

}
