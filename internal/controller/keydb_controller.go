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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const keydbFinalizer = "keydb.keydb/finalizer"

// KeydbReconciler reconciles a Keydb object
type KeydbReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=keydb.keydb,resources=keydbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keydb.keydb,resources=keydbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=keydb.keydb,resources=keydbs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

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
	logger := log.FromContext(ctx)

	var keydb keydbv1.Keydb
	if err := r.Get(ctx, req.NamespacedName, &keydb); err != nil {
		logger.Error(err, "unable to fetch Keydb")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !keydb.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&keydb, keydbFinalizer) {
			// finalizer logic
			controllerutil.RemoveFinalizer(&keydb, keydbFinalizer)
			if err := r.Update(ctx, &keydb); err != nil {
				logger.Error(err, "failed to remove finalizer from keydb")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Ensure finalizer is present
	if !controllerutil.ContainsFinalizer(&keydb, keydbFinalizer) {
		controllerutil.AddFinalizer(&keydb, keydbFinalizer)
		if err := r.Update(ctx, &keydb); err != nil {
			logger.Error(err, "failed to add finalizer to keydb")
			return ctrl.Result{}, err
		}
	}

	// inside your Reconcile after you’ve fetched Keydb CR
	cmList, err := k8sresources.GenerateKeydbConfigMap(&keydb, r.Scheme) // returns []*corev1.ConfigMap
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, cm := range cmList {
		if err := k8sresources.ApplyResource(ctx, r.Client, r.Scheme, &keydb, cm, logger); err != nil {
			return ctrl.Result{}, err
		}
	}

	secret := k8sresources.GenerateSecret(&keydb, r.Scheme, r.Client)
	var secretHash string
	if secret != nil {
		secretHash = k8sresources.HashSecret(secret)
		if err := k8sresources.ApplyResource(ctx, r.Client, r.Scheme, &keydb, secret, logger); err != nil {
			return ctrl.Result{}, err
		}
	} else if keydb.Spec.PasswordSecret != nil {
		secretHash = "custom-" + keydb.Spec.PasswordSecret.Name
	}

	// Now statefulset: inject hashes as podTemplate annotations
	sts := k8sresources.GenerateStatefulSet(&keydb, r.Scheme)
	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = map[string]string{}
	}
	sts.Spec.Template.Annotations["checksum/secret"] = secretHash

	if err := k8sresources.ApplyResource(ctx, r.Client, r.Scheme, &keydb, sts, logger); err != nil {
		return ctrl.Result{}, err
	}

	//services
	svcList, err := k8sresources.GenerateService(&keydb, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, svc := range svcList {
		if err := k8sresources.ApplyResource(ctx, r.Client, r.Scheme, &keydb, svc, logger); err != nil {
			return ctrl.Result{}, err
		}
	}

	// ServiceAccount
	if err := k8sresources.ApplyResource(ctx, r.Client, r.Scheme, &keydb, k8sresources.GenerateServiceAccount(&keydb, r.Scheme), logger); err != nil {
		return ctrl.Result{}, err
	}

	// Pod Disruption Budget
	pdb := k8sresources.GeneratePodDisruptionBudget(&keydb, r.Scheme)
	if err := k8sresources.ApplyResource(ctx, r.Client, r.Scheme, &keydb, pdb, logger); err != nil {
		logger.Error(err, "failed to create/update PodDisruptionBudget")
		// Don't fail reconciliation if PDB fails, but log it
	}

	// Get the current StatefulSet to update status
	var currentSts appsv1.StatefulSet
	stsKey := types.NamespacedName{
		Name:      keydb.Name,
		Namespace: keydb.Namespace,
	}
	if err := r.Get(ctx, stsKey, &currentSts); err != nil {
		logger.V(1).Info("StatefulSet not found yet, will update status on next reconcile", "error", err)
	} else {
		// Update status based on StatefulSet
		if err := r.updateStatus(ctx, &keydb, &currentSts); err != nil {
			logger.Error(err, "failed to update status")
		}
	}

	logger.Info("reconcile cycle completed successfully")
	return ctrl.Result{}, nil
}

// updateStatus updates the KeyDB status based on the StatefulSet status
func (r *KeydbReconciler) updateStatus(ctx context.Context, keydb *keydbv1.Keydb, sts *appsv1.StatefulSet) error {
	logger := log.FromContext(ctx)

	// Get current statefulset status
	readyReplicas := int32(0)
	currentReplicas := int32(0)
	if sts != nil {
		readyReplicas = sts.Status.ReadyReplicas
		currentReplicas = sts.Status.Replicas
	}

	// Update status
	keydb.Status.ReadyReplicas = readyReplicas
	keydb.Status.CurrentReplicas = currentReplicas
	keydb.Status.ObservedGeneration = keydb.Generation
	now := metav1.Now()
	keydb.Status.LastUpdateTime = &now

	// Determine phase
	desiredReplicas := int32(1)
	if keydb.Spec.Replicas != nil {
		desiredReplicas = *keydb.Spec.Replicas
	}

	var phase string
	if currentReplicas == 0 {
		phase = "Pending"
	} else if readyReplicas < desiredReplicas {
		phase = "Scaling"
	} else if readyReplicas == desiredReplicas && currentReplicas == desiredReplicas {
		phase = "Running"
	} else {
		phase = "Unknown"
	}
	keydb.Status.Phase = phase

	// Update conditions
	r.updateConditions(keydb, readyReplicas, desiredReplicas, currentReplicas)

	// Update replica status
	r.updateReplicaStatus(ctx, keydb, sts)

	if err := r.Status().Update(ctx, keydb); err != nil {
		logger.Error(err, "failed to update KeyDB status")
		return err
	}

	return nil
}

// updateConditions updates the conditions in the KeyDB status
func (r *KeydbReconciler) updateConditions(keydb *keydbv1.Keydb, readyReplicas, desiredReplicas, currentReplicas int32) {
	// Ready condition
	readyCondition := metav1.Condition{
		Type:               keydbv1.ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: keydb.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if readyReplicas == desiredReplicas && currentReplicas == desiredReplicas {
		readyCondition.Status = metav1.ConditionTrue
		readyCondition.Reason = keydbv1.ReasonAllReplicasReady
		readyCondition.Message = fmt.Sprintf("All %d replicas are ready", desiredReplicas)
	} else {
		readyCondition.Reason = keydbv1.ReasonSomeReplicasNotReady
		readyCondition.Message = fmt.Sprintf("%d/%d replicas are ready", readyReplicas, desiredReplicas)
	}

	meta.SetStatusCondition(&keydb.Status.Conditions, readyCondition)

	// Progressing condition
	progressingCondition := metav1.Condition{
		Type:               keydbv1.ConditionTypeProgressing,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: keydb.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if currentReplicas != desiredReplicas {
		progressingCondition.Status = metav1.ConditionTrue
		if currentReplicas < desiredReplicas {
			progressingCondition.Reason = keydbv1.ReasonScalingUp
			progressingCondition.Message = fmt.Sprintf("Scaling up from %d to %d replicas", currentReplicas, desiredReplicas)
		} else {
			progressingCondition.Reason = keydbv1.ReasonScalingDown
			progressingCondition.Message = fmt.Sprintf("Scaling down from %d to %d replicas", currentReplicas, desiredReplicas)
		}
	} else {
		progressingCondition.Reason = keydbv1.ReasonReconcileSuccess
		progressingCondition.Message = "Reconciliation successful"
	}

	meta.SetStatusCondition(&keydb.Status.Conditions, progressingCondition)
}

// updateReplicaStatus updates the replica status in the KeyDB status
func (r *KeydbReconciler) updateReplicaStatus(ctx context.Context, keydb *keydbv1.Keydb, sts *appsv1.StatefulSet) {
	if sts == nil {
		return
	}

	// Get pod list
	podList := &corev1.PodList{}
	labels := map[string]string{"apps": keydb.Name}
	listOpts := []client.ListOption{
		client.InNamespace(keydb.Namespace),
		client.MatchingLabels(labels),
	}

	if err := r.List(ctx, podList, listOpts...); err != nil {
		return
	}

	var ready, notReady, failed []string

	for _, pod := range podList.Items {
		podName := pod.Name
		if pod.DeletionTimestamp != nil {
			continue
		}

		switch {
		case pod.Status.Phase == corev1.PodFailed:
			failed = append(failed, podName)
		case pod.Status.Phase == corev1.PodRunning && isPodReady(&pod):
			ready = append(ready, podName)
		default:
			notReady = append(notReady, podName)
		}
	}

	keydb.Status.Replicas = keydbv1.ReplicaStatus{
		Ready:    ready,
		NotReady: notReady,
		Failed:   failed,
	}
}

// isPodReady checks if a pod is ready
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeydbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&keydbv1.Keydb{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ServiceAccount{}).
		Named("keydb").
		Complete(r)
}
