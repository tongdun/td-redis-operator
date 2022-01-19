package standalone

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"td-redis-operator/pkg/apis/cache/v1alpha1"
)

var (
	statefulSetGVK     = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
	redisStandaloneGVK = v1alpha1.SchemeGroupVersion.WithKind("RedisStandalone")
)

// ConfigMapData used to render configmap
type ConfigMapData struct {
	v1alpha1.RedisStandalone

	Extra ConfigMapExtraData
}

// ConfigMapExtraData defines data used by template but not in API
type ConfigMapExtraData struct {
	Secret string
}

func (c *Controller) createStatefulSet(p *v1alpha1.RedisStandalone) (*appsv1.StatefulSet, error) {
	sts := appsv1.StatefulSet{}
	if err := c.statefulSetTemp.Execute(p, &sts); err != nil {
		klog.Errorf("can't render statefulset template for %v", p.Name)
		return nil, err
	}

	sts.OwnerReferences = append(sts.OwnerReferences, *metav1.NewControllerRef(p, redisStandaloneGVK))

	return c.kubeClient.AppsV1().StatefulSets(sts.Namespace).Create(context.TODO(), &sts, metav1.CreateOptions{})
}

func (c *Controller) createService(p *v1alpha1.RedisStandalone) (*corev1.Service, error) {
	svc := corev1.Service{}
	if err := c.serviceTemp.Execute(p, &svc); err != nil {
		klog.Errorf("can't render service template for %v", p.Name)
		return nil, err
	}

	svc.OwnerReferences = append(svc.OwnerReferences, *metav1.NewControllerRef(p, redisStandaloneGVK))

	return c.kubeClient.CoreV1().Services(svc.Namespace).Create(context.TODO(), &svc, metav1.CreateOptions{})
}

func (c *Controller) createConfigMap(p *v1alpha1.RedisStandalone) (*corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{}
	d := ConfigMapData{
		RedisStandalone: *p,
		/*Extra: ConfigMapExtraData{
			Secret: c.redisSecret,
		},*/
	}
	if err := c.configMapTemp.Execute(&d, &cm); err != nil {
		klog.Errorf("can't render configmap template for %v", p.Name)
		return nil, err
	}

	cm.OwnerReferences = append(cm.OwnerReferences, *metav1.NewControllerRef(p, redisStandaloneGVK))

	return c.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})
}

func (c *Controller) tryUpdateStatefulSet(sts *appsv1.StatefulSet, p *v1alpha1.RedisStandalone) (*appsv1.StatefulSet, error) {
	stsNew := appsv1.StatefulSet{}
	if err := c.statefulSetTemp.Execute(p, &stsNew); err != nil {
		klog.Errorf("can't render statefulset template for %v", p.Name)
		return nil, err
	}

	stsNew.OwnerReferences = append(stsNew.OwnerReferences, *metav1.NewControllerRef(p, redisStandaloneGVK))
	if apiequality.Semantic.DeepEqual(sts.Labels, stsNew.Labels) &&
		apiequality.Semantic.DeepEqual(sts.Annotations, stsNew.Annotations) &&
		apiequality.Semantic.DeepEqual(sts.OwnerReferences, stsNew.OwnerReferences) {
		return sts, nil
	}

	klog.Infof("Update statefulset of redis standalone %v", p.Name)

	nsts := sts.DeepCopy()
	nsts.Labels = stsNew.Labels
	nsts.Annotations = stsNew.Annotations
	nsts.OwnerReferences = stsNew.OwnerReferences

	if configHashChanged(sts.Annotations, stsNew.Annotations) {
		nsts.Spec = stsNew.Spec
	}

	return c.kubeClient.AppsV1().StatefulSets(sts.Namespace).Update(context.TODO(), nsts, metav1.UpdateOptions{})
}

func (c *Controller) tryUpdateService(svc *corev1.Service, p *v1alpha1.RedisStandalone) (*corev1.Service, error) {
	svcNew := corev1.Service{}
	if err := c.serviceTemp.Execute(p, &svcNew); err != nil {
		klog.Errorf("can't render service template for %v", p.Name)
		return nil, err
	}

	svcNew.OwnerReferences = append(svcNew.OwnerReferences, *metav1.NewControllerRef(p, redisStandaloneGVK))
	if apiequality.Semantic.DeepEqual(svc.Labels, svcNew.Labels) &&
		apiequality.Semantic.DeepEqual(svc.Annotations, svcNew.Annotations) &&
		apiequality.Semantic.DeepEqual(svc.OwnerReferences, svcNew.OwnerReferences) {
		return svc, nil
	}

	klog.Infof("Update service of redis standalone %v", p.Name)

	nsvc := svc.DeepCopy()
	nsvc.Labels = svcNew.Labels
	nsvc.Annotations = svcNew.Annotations
	nsvc.OwnerReferences = svcNew.OwnerReferences

	return c.kubeClient.CoreV1().Services(svc.Namespace).Update(context.TODO(), nsvc, metav1.UpdateOptions{})
}

func (c *Controller) tryUpdateConfigMap(cm *corev1.ConfigMap, p *v1alpha1.RedisStandalone) (*corev1.ConfigMap, error) {
	cmNew := corev1.ConfigMap{}
	d := ConfigMapData{
		RedisStandalone: *p,
		Extra: ConfigMapExtraData{
			Secret: c.redisSecret,
		},
	}
	if err := c.configMapTemp.Execute(&d, &cmNew); err != nil {
		klog.Errorf("can't render configmap template for %v", p.Name)
		return nil, err
	}

	cmNew.OwnerReferences = append(cmNew.OwnerReferences, *metav1.NewControllerRef(p, redisStandaloneGVK))
	if apiequality.Semantic.DeepEqual(cm.Data, cmNew.Data) &&
		apiequality.Semantic.DeepEqual(cm.Labels, cmNew.Labels) &&
		apiequality.Semantic.DeepEqual(cm.Annotations, cmNew.Annotations) &&
		apiequality.Semantic.DeepEqual(cm.OwnerReferences, cmNew.OwnerReferences) {
		return cm, nil
	}

	klog.Infof("Update configmap of redis standalone %v", p.Name)

	ncm := cm.DeepCopy()
	ncm.Data = cmNew.Data
	ncm.Labels = cmNew.Labels
	ncm.Annotations = cmNew.Annotations
	ncm.OwnerReferences = cmNew.OwnerReferences

	return c.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Update(context.TODO(), ncm, metav1.UpdateOptions{})
}

func (c *Controller) getRedisStandaloneFromPod(pod *corev1.Pod) *v1alpha1.RedisStandalone {
	ref := metav1.GetControllerOf(pod)
	if ref == nil {
		// No controller owns this Pod.
		return nil
	}

	if ref.Kind != statefulSetGVK.Kind {
		// Not a pod owned by a stateful set.
		return nil
	}

	sts, err := c.stsLister.StatefulSets(pod.Namespace).Get(ref.Name)
	if err != nil || sts.UID != ref.UID {
		klog.V(4).Infof("Cannot get statefulset %q for pod %q: %v", ref.Name, pod.Name, err)
		return nil
	}

	// Now find the Deployment that owns that ReplicaSet.
	ref = metav1.GetControllerOf(sts)
	if ref == nil {
		return nil
	}

	if ref.Kind != redisStandaloneGVK.Kind {
		return nil
	}

	mp, err := c.redisStandaloneLister.RedisStandalones(pod.Namespace).Get(ref.Name)
	if err != nil || mp.UID != ref.UID {
		klog.V(4).Infof("Cannot get redis standalone %q for pod %q: %v", ref.Name, pod.Name, err)
		return nil
	}

	return mp
}

func (c *Controller) getRedisStandaloneFromEndpoints(ep *corev1.Endpoints) *v1alpha1.RedisStandalone {
	mp, err := c.redisStandaloneLister.RedisStandalones(ep.Namespace).Get(ep.Name)
	if err != nil {
		klog.V(4).Infof("Cannot get redis standalone %q: %v", ep.Name, err)
		return nil
	}
	return mp
}

func configHashChanged(a, b map[string]string) bool {
	aAnno, aOK := a[v1alpha1.ConfigHashAnnotation]
	bAnno, bOK := b[v1alpha1.ConfigHashAnnotation]

	if !aOK && !bOK {
		return false
	}

	if aOK && bOK && aAnno == bAnno {
		return false
	}

	return true
}
