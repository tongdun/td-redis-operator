package svc

import (
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

func parseSelector(v string) (labels.Selector, error) {
	m := map[string]string{}
	ss := strings.Split(v, ",")

	for _, s := range ss {
		kv := strings.Split(s, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("can't parse kv %v", s)
		}

		k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])

		if _, ok := m[k]; ok {
			return nil, fmt.Errorf("duplicate key %v", k)
		}

		m[k] = v
	}

	return labels.SelectorFromSet(labels.Set(m)), nil
}

func podToEndpointAddress(pod *v1.Pod) *v1.EndpointAddress {
	return &v1.EndpointAddress{
		IP:       pod.Status.PodIP,
		NodeName: &pod.Spec.NodeName,
		TargetRef: &v1.ObjectReference{
			Kind:            "Pod",
			Namespace:       pod.Namespace,
			Name:            pod.Name,
			UID:             pod.UID,
			ResourceVersion: pod.ResourceVersion,
		}}
}

func endpointChanged(pod1, pod2 *v1.Pod) bool {
	endpointAddress1 := podToEndpointAddress(pod1)
	endpointAddress2 := podToEndpointAddress(pod2)

	endpointAddress1.TargetRef.ResourceVersion = ""
	endpointAddress2.TargetRef.ResourceVersion = ""

	return !reflect.DeepEqual(endpointAddress1, endpointAddress2)
}

type addressKey struct {
	ip  string
	uid types.UID
}

func getAddressKey(addr *corev1.EndpointAddress) addressKey {
	key := addressKey{
		ip: addr.IP,
	}
	if addr.TargetRef != nil {
		key.uid = addr.TargetRef.UID
	}

	return key
}
