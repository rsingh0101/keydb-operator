package k8sresources

import (
	keydbv1 "github.com/rsingh0101/keydb-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GenerateServiceAccount(k *keydbv1.Keydb, scheme *runtime.Scheme) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name,
			Namespace: k.Namespace,
			Labels:    map[string]string{"apps": k.Name},
		},
	}

	_ = ctrl.SetControllerReference(k, sa, scheme)
	return sa
}
