package k8sresources

import (
	"net"
	"strings"

	keydbv1 "github.com/rsingh0101/keydb-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GenerateStatefulSet(k *keydbv1.Keydb, scheme *runtime.Scheme) *appsv1.StatefulSet {
	labels := map[string]string{
		"apps": k.Name,
	}
	storageClassName := k.Spec.Persistence.StorageClassName
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
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data-" + k.Name + "-pvc",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(k.Spec.Persistence.Size),
							},
						},
						StorageClassName: &storageClassName,
					},
				},
			},
			ServiceName: k.Name + "-headless",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: k.Name,
					Containers: []corev1.Container{
						{

							SecurityContext: &corev1.SecurityContext{
								RunAsNonRoot: &[]bool{true}[0],
								RunAsUser:    &[]int64{1001}[0],
								RunAsGroup:   &[]int64{1001}[0],
							},
							Name:  "keydb",
							Image: k.Spec.Image,
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"-ec",
								`. /opt/bitnami/scripts/keydb-env.sh && \
                 args=("/opt/bitnami/keydb/etc/keydb.conf") \
                 args+=("--requirepass "$KEYDB_PASSWORD") \
                 args+=("--masterauth "$KEYDB_MASTER_PASSWORD") \
                 args+=("--replicaof keydb-0 6379") \
				 exec keydb-server "${args[@]}"`,
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 6379,
									Name:          "keydb",
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"/opt/bitnami/scripts/health/ping_liveness_local_and_master.sh 5",
										},
									},
								},
								InitialDelaySeconds: 20,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    5,
								TimeoutSeconds:      5,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"/opt/bitnami/scripts/health/ping_readiness_local_and_master.sh 5",
										},
									},
								},
								InitialDelaySeconds: 20,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    5,
								TimeoutSeconds:      5,
							},
							Env: []corev1.EnvVar{
								{
									Name: "KEYDB_PASSWORD_FILE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: k.Name + "-secret",
											},
											Key: "password",
										},
									},
								},
								{
									Name: "KEYDB_MASTER_PASSWORD_FILE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: k.Name + "-secret",
											},
											Key: "password",
										},
									},
								},
								{
									Name:  "KEYDB_PORT_NUMBER",
									Value: "6379",
								},
								{
									Name:  "BITNAMI_DEBUG",
									Value: "false",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      k.Name + "-config",
									MountPath: "/opt/bitnami/keydb/etc/",
								},
								{
									Name:      "data-" + k.Name + "-pvc",
									MountPath: "/bitnami/keydb/data/",
								},
								{
									Name:      k.Name + "-healthz",
									MountPath: "/opt/bitnami/scripts/health",
									ReadOnly:  true,
								},
								{
									Name:      k.Name + "-secret",
									MountPath: "/opt/bitnami/keydb/secrets",
								},
								{
									Name:      "empty-dir",
									MountPath: "/tmp",
									SubPath:   "tmp-dir",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data-" + k.Name + "-pvc",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "data-" + k.Name + "-pvc",
								},
							},
						},
						{
							Name: k.Name + "-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: k.Name + "-config",
									},
								},
							},
						},
						{
							Name: k.Name + "-healthz",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: k.Name + "-healthz",
									},
								},
							},
						},
						{
							Name: k.Name + "-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: k.Name + "-secret",
									Items: []corev1.KeyToPath{
										{
											Key:  "password",
											Path: "password",
										},
									},
								},
							},
						},
						{
							Name: "empty-dir",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	_ = ctrl.SetControllerReference(k, sts, scheme)
	return sts
}

// normalizeFQDN appends .svc.cluster.local if the user provided a short service name.
func NormalizeFQDN(h string) string {
	if net.ParseIP(h) != nil {
		return h
	}
	// already looks like an FQDN (has at least 2 dots: e.g., svc.cluster.local)
	if strings.Count(h, ".") >= 2 {
		return h
	}
	// allow "svc.ns" short form -> "svc.ns.svc.cluster.local"
	if strings.Count(h, ".") == 1 {
		return h + ".svc.cluster.local"
	}
	// single token (service in same ns) -> best-effort suffix
	return h + ".svc.cluster.local"
}
