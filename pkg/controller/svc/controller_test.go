package svc

import (
	"fmt"
	"testing"
	"time"

	"github.com/liubog2008/tester/pkg/data"
	"github.com/liubog2008/tester/pkg/tester"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
	"k8s.io/client-go/util/workqueue"
	"td-redis-operator/pkg/controller"
	utiltesting "td-redis-operator/pkg/utils/testing"
)

func TestNewService(t *testing.T) {
	tester.Test(t, testNewService)
}

var (
	noResyncPeriodFunc = func() time.Duration { return 0 }
	defaultTimeout     = 5 * time.Second

	serviceGVR = corev1.SchemeGroupVersion.WithResource("services")
	// podGVR       = corev1.SchemeGroupVersion.WithResource("pods")
	endpointsGVR = corev1.SchemeGroupVersion.WithResource("endpoints")
)

type fakeController struct {
	fake       *Controller
	kubeClient *kubefake.Clientset
	starter    chan struct{}
	finisher   chan struct{}
}

func fakeConcilerFactory(starter <-chan struct{}, finisher chan<- struct{}, f controller.ReconcilerFactory) controller.ReconcilerFactory {
	return func(queue workqueue.RateLimitingInterface, syncer controller.Syncer) Reconciler {
		nf := f(queue, syncer)
		return func() (bool, error) {
			<-starter
			quit, err := nf()
			finisher <- struct{}{}
			return quit, err
		}
	}
}

func (c *fakeController) Run(stopCh <-chan struct{}) {
	f := c.fake.reconcilerFactory
	c.fake.reconcilerFactory = fakeConcilerFactory(c.starter, c.finisher, f)

	c.fake.Run(1, stopCh)
}

func runFakeController(kubeClient *kubefake.Clientset, stopCh <-chan struct{}) *fakeController {
	fakeKubeInformers := kubeinformers.NewSharedInformerFactory(kubeClient, noResyncPeriodFunc())

	fakePodInformer := fakeKubeInformers.Core().V1().Pods()
	fakeServiceInformer := fakeKubeInformers.Core().V1().Services()
	fakeEndpointsInformer := fakeKubeInformers.Core().V1().Endpoints()

	fc := &fakeController{
		fake: NewController(&ControllerOptions{
			KubeClient:        kubeClient,
			PodInformer:       fakePodInformer,
			ServiceInformer:   fakeServiceInformer,
			EndpointsInformer: fakeEndpointsInformer,
		}),
		starter:    make(chan struct{}),
		finisher:   make(chan struct{}),
		kubeClient: kubeClient,
	}

	go fakeKubeInformers.Start(stopCh)

	go fc.Run(stopCh)

	return fc
}

func (c *fakeController) syncOnce(timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	c.starter <- struct{}{}
	select {
	case <-c.finisher:
	case <-timer.C:
		return fmt.Errorf("sync timeout")
	}

	return nil
}

func (c *fakeController) sync(timeout time.Duration, times int) error {
	for i := 0; i < times; i++ {
		if err := c.syncOnce(timeout); err != nil {
			return err
		}
	}

	return nil
}

func (c *fakeController) MustSyncOnce(t *testing.T) {
	err := c.syncOnce(defaultTimeout)
	require.NoError(t, err, "sync once timeout")
}

func (c *fakeController) MustSync(t *testing.T, times int) {
	err := c.sync(defaultTimeout, times)
	require.NoError(t, err, "sync %v times timeout", times)
}

type CaseNewService struct {
	Pods []*corev1.Pod `json:"pods"`

	Service           *corev1.Service   `json:"service"`
	ExpectedEndpoints *corev1.Endpoints `json:"expectedEndpoints"`
}

func testNewService(t *testing.T, tc data.TestCase) {
	c := CaseNewService{}
	require.NoError(t, tc.Unmarshal(&c))

	stopCh := make(chan struct{})

	kubeStore := []runtime.Object{}
	for _, p := range c.Pods {
		kubeStore = append(kubeStore, p)
	}

	kubeClient := kubefake.NewSimpleClientset(kubeStore...)
	fc := runFakeController(kubeClient, stopCh)
	_, err := fc.kubeClient.CoreV1().Services(c.Service.Namespace).Create(c.Service)
	require.NoError(t, err, tc.Description(), "create service error")
	fc.MustSyncOnce(t)
	actions := fc.kubeClient.Actions()
	expected := []kubetesting.Action{
		kubetesting.NewCreateAction(serviceGVR, c.Service.Namespace, c.Service),
		kubetesting.NewCreateAction(endpointsGVR, c.ExpectedEndpoints.Namespace, c.ExpectedEndpoints),
	}

	utiltesting.AssertWriteActions(t, expected, actions)
}
