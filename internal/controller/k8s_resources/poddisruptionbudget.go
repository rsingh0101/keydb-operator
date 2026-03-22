package k8sresources

import (
	keydbv1 "github.com/rsingh0101/keydb-operator/api/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GeneratePodDisruptionBudget(k *keydbv1.Keydb, scheme *runtime.Scheme) *policyv1.PodDisruptionBudget {
	labels := map[string]string{
		"apps": k.Name,
	}

	replicas := int32(1)
	if k.Spec.Replicas != nil {
		replicas = *k.Spec.Replicas
	}

	// Calculate minAvailable: at least 1, or maintain quorum for master-master
	var minAvailable intstr.IntOrString
	if replicas == 1 {
		minAvailable = intstr.FromInt32(1)
	} else {
		// For master-master replication, maintain quorum (majority)
		// For 3 replicas: minAvailable = 2 (quorum)
		// For 5 replicas: minAvailable = 3 (quorum)
		if k.Spec.Replication.Mode == "master-master" {
			minAvailableInt := (replicas / 2) + 1
			minAvailable = intstr.FromInt32(minAvailableInt)
		} else {
			// For other modes, allow at least 1 pod to be unavailable
			minAvailableInt := replicas - 1
			if minAvailableInt < 1 {
				minAvailableInt = 1
			}
			minAvailable = intstr.FromInt32(minAvailableInt)
		}
	}

	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name + "-pdb",
			Namespace: k.Namespace,
			Labels:    labels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}

	_ = ctrl.SetControllerReference(k, pdb, scheme)
	return pdb
}
