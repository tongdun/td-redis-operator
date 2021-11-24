// Package mysqlproxy defines controller to manage mysql proxy
package mysqlproxy

import (
	"fmt"

	"redis-priv-operator/pkg/client/clientset"
	tdbinformers "redis-priv-operator/pkg/client/informers/tdb/v1alpha1"
	tdblisters "redis-priv-operator/pkg/client/listers/tdb/v1alpha1"
	"redis-priv-operator/pkg/controller"
	"redis-priv-operator/pkg/template"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// ControllerOptions defines options for mysql proxy controller
type ControllerOptions struct {
	// KubeClient defines interface of raw kubernetes API
	KubeClient kubernetes.Interface

	// ExtClient defines interface of CR extension API
	ExtClient clientset.Interface

	// StatefulSetTemplate defines template of mysql proxy
	StatefulSetTemplate *template.Template

	// ServiceTemplate defines template of mysql proxy service
	ServiceTemplate *template.Template

	// ConfigMapTemplate defines template of mysql proxy configmap
	ConfigMapTemplate *template.Template

	// PodInformer defines informer of pod
	PodInformer coreinformers.PodInformer

	// EndpointsInformer defines informer of endpoints
	EndpointsInformer coreinformers.EndpointsInformer

	// ServiceInformer defines informer of service
	ServiceInformer coreinformers.ServiceInformer

	// ConfigMapInformer defines informer of configmap
	ConfigMapInformer coreinformers.ConfigMapInformer

	// StatefulSetInformer defines informer of statefulset
	StatefulSetInformer appsinformers.StatefulSetInformer

	// MysqlProxyInfromer defines informer of mysql proxy
	MysqlProxyInformer tdbinformers.MysqlProxyInformer

	// MysqlSecret defines secret for mysql proxy
	MysqlSecret string
}

// Controller defines controller to manage mysql proxy
type Controller struct {
	kubeClient kubernetes.Interface
	extClient  clientset.Interface

	informersSynced []cache.InformerSynced

	queue workqueue.RateLimitingInterface

	mysqlProxyLister tdblisters.MysqlProxyLister

	stsLister appslisters.StatefulSetLister
	svcLister corelisters.ServiceLister
	podLister corelisters.PodLister
	cmLister  corelisters.ConfigMapLister
	epLister  corelisters.EndpointsLister

	statefulSetTemp *template.Template
	serviceTemp     *template.Template
	configMapTemp   *template.Template

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	reconcilerFactory controller.ReconcilerFactory

	mysqlSecret string
}

// NewController returns a tdb controller
func NewController(opt *ControllerOptions) *Controller {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: opt.KubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "mysql-proxy-operator"})

	c := &Controller{
		kubeClient: opt.KubeClient,
		extClient:  opt.ExtClient,
		informersSynced: []cache.InformerSynced{
			opt.PodInformer.Informer().HasSynced,
			opt.ServiceInformer.Informer().HasSynced,
			opt.ConfigMapInformer.Informer().HasSynced,
			opt.StatefulSetInformer.Informer().HasSynced,
			opt.MysqlProxyInformer.Informer().HasSynced,
		},
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mysqlproxy"),
		mysqlProxyLister: opt.MysqlProxyInformer.Lister(),
		stsLister:        opt.StatefulSetInformer.Lister(),
		podLister:        opt.PodInformer.Lister(),
		svcLister:        opt.ServiceInformer.Lister(),
		cmLister:         opt.ConfigMapInformer.Lister(),
		epLister:         opt.EndpointsInformer.Lister(),

		statefulSetTemp: opt.StatefulSetTemplate,
		serviceTemp:     opt.ServiceTemplate,
		configMapTemp:   opt.ConfigMapTemplate,

		eventBroadcaster:  broadcaster,
		eventRecorder:     recorder,
		reconcilerFactory: controller.RateLimitingReconciler,

		mysqlSecret: opt.MysqlSecret,
	}

	opt.PodInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addPod,
		UpdateFunc: c.updatePod,
		DeleteFunc: c.deletePod,
	})

	opt.EndpointsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addEndpoints,
		UpdateFunc: c.updateEndpoints,
		DeleteFunc: c.deleteEndpoints,
	})

	opt.MysqlProxyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addMysqlProxy,
		UpdateFunc: c.updateMysqlProxy,
		DeleteFunc: c.deleteMysqlProxy,
	})

	return c
}

func (c *Controller) addMysqlProxy(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	c.queue.Add(key)
}

func (c *Controller) updateMysqlProxy(old, cur interface{}) {
	c.addMysqlProxy(cur)
}

func (c *Controller) deleteMysqlProxy(obj interface{}) {
	c.addMysqlProxy(obj)
}

func (c *Controller) addEndpoints(obj interface{}) {
	ep := obj.(*corev1.Endpoints)
	mp := c.getMysqlProxyFromEndpoints(ep)
	if mp == nil {
		return
	}
	c.addMysqlProxy(mp)
}

func (c *Controller) updateEndpoints(old, cur interface{}) {
	c.addEndpoints(cur)
}

func (c *Controller) deleteEndpoints(obj interface{}) {
	c.addEndpoints(obj)
	if ep, ok := obj.(*corev1.Endpoints); ok {
		c.addEndpoints(ep)
		return
	}
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
		return
	}
	ep, ok := tombstone.Obj.(*corev1.Endpoints)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a Endpoints: %#v", obj))
		return
	}

	c.addEndpoints(ep)
}

func (c *Controller) addPod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	mp := c.getMysqlProxyFromPod(pod)
	if mp == nil {
		return
	}
	c.addMysqlProxy(mp)
}

func (c *Controller) updatePod(old, cur interface{}) {
	oldPod := old.(*corev1.Pod)
	newPod := cur.(*corev1.Pod)
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		return
	}
	// TODO(bo.liub): check whether pod status change?
	c.addPod(newPod)
}

func (c *Controller) deletePod(obj interface{}) {
	if pod, ok := obj.(*corev1.Pod); ok {
		c.addPod(pod)
		return
	}
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
		return
	}
	pod, ok := tombstone.Obj.(*corev1.Pod)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a Pod: %#v", obj))
		return
	}

	c.addPod(pod)
}

// Run will start the controller
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting mysqlproxy controller")
	defer klog.Infof("Shutting down mysqlproxy controller")

	if !cache.WaitForCacheSync(stopCh, c.informersSynced...) {
		utilruntime.HandleError(fmt.Errorf("unable to sync caches for mysqlproxy controller"))
		return
	}

	klog.Infof("Cache has been synced")

	for i := 0; i < workers; i++ {
		controller.WaitUntil("mysqlproxy", c.reconcilerFactory(c.queue, c.syncMysqlProxy), stopCh)
	}

	klog.Infof("mysqlproxy controller is working")
	<-stopCh
}
