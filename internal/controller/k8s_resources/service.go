package k8sresources

import (
	keydbv1 "github.com/rsingh0101/keydb-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GenerateService(k *keydbv1.Keydb, scheme *runtime.Scheme) ([]*corev1.Service, error) {
	labels := map[string]string{
		"apps": k.Name,
	}

	// ClusterIP Service (default access point)
	clusterSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name + "-svc",
			Namespace: k.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			PublishNotReadyAddresses: true,
			ClusterIP:                corev1.ClusterIPNone,
			Selector:                 labels,
			Ports: []corev1.ServicePort{
				{
					Name: "redis",
					Port: 6379,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	if err := ctrl.SetControllerReference(k, clusterSvc, scheme); err != nil {
		return nil, err
	}

	// Headless Service (for StatefulSet pod DNS)
	headlessSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name + "-headless",
			Namespace: k.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name: "redis",
					Port: 6379,
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(k, headlessSvc, scheme); err != nil {
		return nil, err
	}

	return []*corev1.Service{clusterSvc, headlessSvc}, nil
}
