package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"redis-priv-operator/pkg/apis/tdb/v1alpha1"
)

var (
	renderedConfigMap = `global
    nbproc 1
    pidfile haproxy.pid
defaults
    mode tcp
    retries 2
    option redispatch
    option abortonclose
    maxconn 4096
    timeout connect 1000ms
    timeout client 10800s
    timeout server 10800s
    log 127.0.0.1 local0 debug

frontend mysql
    bind *:3306
    mode tcp
    default_backend mysqlservers
backend mysqlservers
    balance leastconn
    option mysql-check user repl:repl.tongdun.CN  post-41
    server master0 192.168.1.1:2311 check inter 2000 rise 1 fall 2
    server master1 192.168.1.2:2311 check backup inter 2000 rise 1 fall 2
`
)

// TODO(bo.liub): chang it
type ConfigMapData struct {
	v1alpha1.MysqlProxy
	Extra ConfigMapExtraData
}

type ConfigMapExtraData struct {
	Secret string
}

func TestConfigMapTempalte(t *testing.T) {
	cases := []struct {
		desc string
		s    string
		p    *ConfigMapData
		cm   *corev1.ConfigMap
	}{
		{
			desc: "new template",
			s:    "/configmap.tmpl",
			p: &ConfigMapData{
				MysqlProxy: v1alpha1.MysqlProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "xxx",
						Namespace: "yyy",
						UID:       "yy",
					},
					Spec: v1alpha1.MysqlProxySpec{
						Mysqls: []v1alpha1.Mysql{
							{
								Name: "aa",
								IP:   "192.168.1.1",
								Port: "2311",
							},
							{
								Name: "bb",
								IP:   "192.168.1.2",
								Port: "2311",
							},
						},
					},
				},
				Extra: ConfigMapExtraData{
					Secret: "repl.tongdun.CN",
				},
			},
			cm: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xxx",
					Namespace: "yyy",
				},
				Data: map[string]string{
					"haproxy.cfg": renderedConfigMap,
				},
			},
		},
	}

	for _, c := range cases {
		templ, err := NewTemplate(c.s)
		require.NoError(t, err, c.desc)

		cm := &corev1.ConfigMap{}
		require.NoError(t, templ.Execute(c.p, cm), c.desc)
		assert.Equal(t, c.cm, cm, c.desc)
	}
}

func TestServiceTempalte(t *testing.T) {
	cases := []struct {
		desc string
		s    string
		p    *v1alpha1.MysqlProxy
		svc  *corev1.Service
	}{
		{
			desc: "new template",
			s:    "/service.tmpl",
			p: &v1alpha1.MysqlProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xx",
					Namespace: "yy",
				},
				Spec: v1alpha1.MysqlProxySpec{
					Suspended: true,
				},
			},
			svc: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xx",
					Namespace: "yy",
					Annotations: map[string]string{
						v1alpha1.SingletonServiceSelectorAnnotation: "app=mysqlproxy,name=xx",
						v1alpha1.SingletonServiceStatusAnnotation:   v1alpha1.SingletonServiceDisabled,
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port:       3306,
							TargetPort: intstr.FromInt(3306),
							Protocol:   corev1.ProtocolTCP,
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		templ, err := NewTemplate(c.s)
		require.NoError(t, err, c.desc)

		svc := &corev1.Service{}
		require.NoError(t, templ.Execute(c.p, svc), c.desc)
		assert.Equal(t, c.svc, svc, c.desc)
	}
}

// TODO(bo.liub): it should be extracted
// Copy from cmd/admin/networkpolicy.go
type NetworkPolicy struct {
	Name string `json:"name"`

	IPBlocks []netv1.IPBlock `json:"ipBlocks,omitempty"`

	Namespace string `json:"-"`

	DenyAll bool `json:"-"`

	Service string
	Port    int32
}

