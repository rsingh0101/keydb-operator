/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	keydbv1 "github.com/rsingh0101/keydb-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// KeydbReconciler reconciles a Keydb object
type KeydbReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=keydb.keydb,resources=keydbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keydb.keydb,resources=keydbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=keydb.keydb,resources=keydbs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Keydb object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *KeydbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var keydb keydbv1.Keydb
	if err := r.Get(ctx, req.NamespacedName, &keydb); err != nil {
		log.Error(err, "unable to fetch Keydb")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	sts := r.generateStatefulSet(&keydb)
	var existingSts appsv1.StatefulSet
	err := r.Get(ctx, types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, &existingSts)
	if err != nil && apierrors.IsNotFound(err) {
		if err := r.Create(ctx, sts); err != nil {
			log.Error(err, "failed to create statefulset")
			return ctrl.Result{}, err
		}
		log.Info("Created Statefulset", "name", sts.Name)
	} else if err != nil {
		log.Error(err, "failed to fetch statefulset")
		return ctrl.Result{}, err
	}
	keydb.Status.Phase = "Running"
	if err := r.Status().Update(ctx, &keydb); err != nil {
		log.Error(err, "failed to update status")
	}
	return ctrl.Result{}, nil
}

func (r *KeydbReconciler) generateStatefulSet(k *keydbv1.Keydb) *appsv1.StatefulSet {
	labels := map[string]string{
		"apps": k.Name,
	}
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name,
			Namespace: k.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: k.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: k.Name + "-svc",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "keydb",
							Image: k.Spec.Image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 6379},
							},
						},
					},
				},
			},
		},
	}

	_ = ctrl.SetControllerReference(k, sts, r.Scheme)
	return sts
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeydbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&keydbv1.Keydb{}).
		Owns(&appsv1.StatefulSet{}).
		Named("keydb").
		Complete(r)
}
