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
	"fmt"

	keydbv1 "github.com/rsingh0101/keydb-operator/api/v1"
	k8sresources "github.com/rsingh0101/keydb-operator/internal/controller/k8s_resources"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	//services
	svc, err := k8sresources.GenerateService(&keydb, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, svc := range svc {
		if err := k8sresources.CreateorUpdateResource(ctx, r.Client, r.Scheme, svc, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	//configmap
	configmap, err := k8sresources.GenerateKeydbConfigMap(&keydb, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, cm := range configmap {
		if err := k8sresources.CreateorUpdateResource(ctx, r.Client, r.Scheme, cm, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	//secret
	if err := k8sresources.CreateorUpdateResource(ctx, r.Client, r.Scheme, k8sresources.GenerateSecret(&keydb, r.Scheme, r.Client), log); err != nil {
		return ctrl.Result{}, err
	}

	//ServiceAccount
	if err := k8sresources.CreateorUpdateResource(ctx, r.Client, r.Scheme, k8sresources.GenerateServiceAccount(&keydb, r.Scheme), log); err != nil {
		return ctrl.Result{}, err
	}
	//statefulset
	if err := k8sresources.CreateorUpdateResource(ctx, r.Client, r.Scheme, k8sresources.GenerateStatefulSet(&keydb, r.Scheme), log); err != nil {
		return ctrl.Result{}, err
	}

	keydb.Status.Phase = "Running"
	// log.Info("Phase updated successfully for keydb", "name", keydb.Name)
	fmt.Printf("\033[32m✅Phase updated successfully for Keydb\n")
	if err := r.Status().Update(ctx, &keydb); err != nil {
		log.Error(err, "failed to update status")
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeydbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&keydbv1.Keydb{}).
		Owns(&appsv1.StatefulSet{}).
		Named("keydb").
		Complete(r)
}
