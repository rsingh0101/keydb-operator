// reconcile_helpers.go
package k8sresources

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ApplyResource(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	owner client.Object,
	desired client.Object,
	log logr.Logger,
) error {
	// Set controller reference
	if err := controllerutil.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}

	// Create an empty object of the same type to use with CreateOrUpdate
	existing := desired.DeepCopyObject().(client.Object)
	existing.SetResourceVersion("")
	existing.SetGeneration(0)
	existing.SetUID("")

	op, err := controllerutil.CreateOrUpdate(ctx, c, existing, func() error {
		// Copy metadata annotations and labels
		existing.SetLabels(desired.GetLabels())

		// Map specific struct fields
		switch e := existing.(type) {
		case *appsv1.StatefulSet:
			d := desired.(*appsv1.StatefulSet)
			e.Spec = d.Spec

			// Copy template annotations that might have been dynamically added (like restartedAt)
			if d.Spec.Template.Annotations != nil {
				if e.Spec.Template.Annotations == nil {
					e.Spec.Template.Annotations = make(map[string]string)
				}
				for k, v := range d.Spec.Template.Annotations {
					e.Spec.Template.Annotations[k] = v
				}
			}

		case *corev1.Service:
			d := desired.(*corev1.Service)
			clusterIP := e.Spec.ClusterIP
			e.Spec = d.Spec
			if clusterIP != "" && clusterIP != "None" {
				e.Spec.ClusterIP = clusterIP
			}

		case *corev1.ConfigMap:
			d := desired.(*corev1.ConfigMap)
			e.Data = d.Data

		case *corev1.Secret:
			d := desired.(*corev1.Secret)
			e.Data = d.Data
			e.StringData = d.StringData

		case *policyv1.PodDisruptionBudget:
			d := desired.(*policyv1.PodDisruptionBudget)
			e.Spec = d.Spec

		case *corev1.ServiceAccount:
			// No spec to copy
		}

		return nil
	})

	if err != nil {
		log.Error(err, "reconcile failed", "kind", desired.GetObjectKind().GroupVersionKind().Kind, "name", desired.GetName())
		return err
	}

	if op != controllerutil.OperationResultNone {
		log.Info("reconciled", "kind", desired.GetObjectKind().GroupVersionKind().Kind, "name", desired.GetName(), "operation", op)
	}

	desired.SetResourceVersion(existing.GetResourceVersion())
	return nil
}
