// Package app defines operator
package app

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog"
	"td-redis-operator/cmd/operator/app/config"
	"td-redis-operator/cmd/operator/app/options"
	"td-redis-operator/pkg/controller/redis/cluster"
	"td-redis-operator/pkg/controller/redis/masterslave"
	"td-redis-operator/pkg/controller/redis/standalone"
	"td-redis-operator/pkg/controller/svc"
)

// NewCommand returns app command
func NewCommand() *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:  "operator",
		Long: "operator for deploy redis resource",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()
			printFlags(cmd.Flags())
			cfg, err := opts.Config()
			if err != nil {
				klog.Fatalf("can't parse options to config: %v", err)
			}
			stopCh := make(chan struct{})
			if err := Run(cfg, stopCh); err != nil {
				klog.Fatalf("run operator failed: %v", err)
			}
		},
	}
	opts.AddFlags(cmd.Flags())

	return cmd
}

// Run runs the operator
func Run(cfg *config.Config, stopCh <-chan struct{}) error {
	svcController := newServiceController(cfg)
	//mysqlProxyController := newMysqlProxyController(cfg)
	//redisStandaloneController := newRedisStandaloneController(cfg)
	redisStandbyController := newRedisStandbyController(cfg)
	redisClusterController := newRedisClusterController(cfg)

	go cfg.KubeInformerFactory.Start(stopCh)

	go cfg.ExtInformerFactory.Start(stopCh)

	go svcController.Run(1, stopCh)

	//go mysqlProxyController.Run(1, stopCh)

	//go redisStandaloneController.Run(1, stopCh)

	go redisStandbyController.Run(1, stopCh)

	go redisClusterController.Run(1, stopCh)

	<-stopCh

	return nil
}

func newServiceController(cfg *config.Config) *svc.Controller {
	opt := &svc.ControllerOptions{
		KubeClient:           cfg.KubeClient,
		RedisClusterInformer: cfg.RedisClusterInformer,
		PodInformer:          cfg.PodInformer,
		ServiceInformer:      cfg.ServiceInformer,
		EndpointsInformer:    cfg.EndpointsInformer,
	}
	c := svc.NewController(opt)

	return c
}

func newRedisStandaloneController(cfg *config.Config) *standalone.Controller {
	opt := &standalone.ControllerOptions{
		KubeClient:              cfg.KubeClient,
		ExtClient:               cfg.ExtClient,
		PodInformer:             cfg.PodInformer,
		ServiceInformer:         cfg.ServiceInformer,
		ConfigMapInformer:       cfg.ConfigMapInformer,
		EndpointsInformer:       cfg.EndpointsInformer,
		StatefulSetInformer:     cfg.StatefulSetInformer,
		RedisStandaloneInformer: cfg.RedisStandaloneInformer,
		StatefulSetTemplate:     cfg.RedisStandaloneStatefulSetTemplate,
		ServiceTemplate:         cfg.RedisStandaloneServiceTemplate,
		ConfigMapTemplate:       cfg.RedisStandaloneConfigMapTemplate,
	}
	c := standalone.NewController(opt)
	return c
}

func newRedisStandbyController(cfg *config.Config) *masterslave.Controller {
	opt := &masterslave.ControllerOptions{
		KubeClient:               cfg.KubeClient,
		ExtClient:                cfg.ExtClient,
		PodInformer:              cfg.PodInformer,
		ServiceInformer:          cfg.ServiceInformer,
		ConfigMapInformer:        cfg.ConfigMapInformer,
		EndpointsInformer:        cfg.EndpointsInformer,
		StatefulSetInformer:      cfg.StatefulSetInformer,
		RedisStandbyInformer:     cfg.RedisStandbyInformer,
		StatefulSetTemplate:      cfg.RedisStandbyStatefulSetTemplate,
		SentiStatefulSetTemplate: cfg.SentinelStatefulSetTemplate,
		ServiceTemplate:          cfg.RedisStandbyServiceTemplate,
		SentiServiceTemplate:     cfg.SentiServiceTemplate,
		ConfigMapTemplate:        cfg.RedisStandbyConfigMapTemplate,
	}
	c := masterslave.NewController(opt)
	return c
}

func newRedisClusterController(cfg *config.Config) *cluster.Controller {
	opt := &cluster.ControllerOptions{
		KubeClient:               cfg.KubeClient,
		ExtClient:                cfg.ExtClient,
		PodInformer:              cfg.PodInformer,
		ServiceInformer:          cfg.ServiceInformer,
		ConfigMapInformer:        cfg.ConfigMapInformer,
		EndpointsInformer:        cfg.EndpointsInformer,
		StatefulSetInformer:      cfg.StatefulSetInformer,
		DeploymentInformer:       cfg.DeploymentInformer,
		PredixyServiceTemplate:   cfg.PredixyServiceTemplate,
		PredixyConfigMapTemplate: cfg.ClusterPredixyConfigMapTemplate,
		PredixyTemplate:          cfg.PredixyTemplate,
		RedisClusterInformer:     cfg.RedisClusterInformer,
		StatefulSetTemplate:      cfg.RedisClusterStatefulSetTemplate,
		ServiceTemplate:          cfg.RedisClusterServiceTemplate,
		ConfigMapTemplate:        cfg.RedisClusterConfigMapTemplate,
	}
	c := cluster.NewController(opt)
	return c
}

func printFlags(fs *pflag.FlagSet) {
	fs.VisitAll(func(f *pflag.Flag) {
		klog.Infof("FLAG: --%v=%v", f.Name, f.Value)
	})
}
