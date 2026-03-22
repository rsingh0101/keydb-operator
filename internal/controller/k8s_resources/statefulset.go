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
	var volumeClaimTemplates []corev1.PersistentVolumeClaim
	if k.Spec.Persistence.Size != "" {
		var scPtr *string
		if storageClassName != "" {
			scPtr = &storageClassName
		}
		volumeClaimTemplates = []corev1.PersistentVolumeClaim{
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
					StorageClassName: scPtr,
				},
			},
		}
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      k.Name + "-config",
			MountPath: "/opt/bitnami/keydb/etc/",
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
	}

	secretName := k.Name + "-secret"
	secretKey := "password"
	if k.Spec.PasswordSecret != nil {
		secretName = k.Spec.PasswordSecret.Name
		secretKey = k.Spec.PasswordSecret.Key
	}

	volumes := []corev1.Volume{
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
					DefaultMode: func() *int32 { m := int32(0755); return &m }(),
				},
			},
		},
		{
			Name: k.Name + "-secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
					Items: []corev1.KeyToPath{
						{
							Key:  secretKey,
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
	}
	// Only add PVC if persistence is enabled
	if k.Spec.Persistence.Enabled {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "data-" + k.Name + "-pvc",
			MountPath: "/bitnami/keydb/data/",
		})

		volumes = append(volumes, corev1.Volume{
			Name: "data-" + k.Name + "-pvc",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "data-" + k.Name + "-pvc",
				},
			},
		})
	} else {
		// use EmptyDir as fallback
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "data-empty-dir",
			MountPath: "/bitnami/keydb/data/",
		})

		volumes = append(volumes, corev1.Volume{
			Name: "data-empty-dir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
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
			VolumeClaimTemplates: volumeClaimTemplates,
			ServiceName:          k.Name + "-headless",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: k.Name,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
						RunAsUser:    &[]int64{1001}[0],
						RunAsGroup:   &[]int64{1001}[0],
						FSGroup:      &[]int64{1001}[0],
					},
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
								`. /opt/bitnami/scripts/keydb-env.sh;if [ -f /opt/bitnami/scripts/health/fix_replication_config.sh ]; then bash /opt/bitnami/scripts/health/fix_replication_config.sh || true; fi;CONFIG_FILE="${KEYDB_MODIFIED_CONFIG:-/opt/bitnami/keydb/etc/keydb.conf}";args=("${CONFIG_FILE}");args+=("--requirepass" "$KEYDB_PASSWORD");args+=("--masterauth" "$KEYDB_MASTER_PASSWORD");exec keydb-server "${args[@]}"`,
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
									Name:  "KEYDB_PASSWORD_FILE",
									Value: "/opt/bitnami/keydb/secrets/password",
								},
								{
									Name:  "KEYDB_MASTER_PASSWORD_FILE",
									Value: "/opt/bitnami/keydb/secrets/password",
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
							VolumeMounts: volumeMounts,
							// Add resources if specified, otherwise use defaults
							Resources: getContainerResources(k),
						},
						{
							SecurityContext: &corev1.SecurityContext{
								RunAsNonRoot: &[]bool{true}[0],
								RunAsUser:    &[]int64{1001}[0],
								RunAsGroup:   &[]int64{1001}[0],
							},
							Name:  "config-reloader",
							Image: k.Spec.Image,
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"-c",
								". /opt/bitnami/scripts/keydb-env.sh; exec bash /opt/bitnami/scripts/health/config_reloader.sh",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "KEYDB_PASSWORD_FILE",
									Value: "/opt/bitnami/keydb/secrets/password",
								},
								{
									Name:  "KEYDB_PORT_NUMBER",
									Value: "6379",
								},
							},
							VolumeMounts: volumeMounts,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("32Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
					Volumes: volumes,
					// Add pod anti-affinity for better distribution
					Affinity: getPodAffinity(labels),
				},
			},
		},
	}

	if k.Spec.Metrics.Enabled {
		metricsImage := "oliver006/redis_exporter:latest"
		if k.Spec.Metrics.Image != "" {
			metricsImage = k.Spec.Metrics.Image
		}
		sts.Spec.Template.Spec.Containers = append(sts.Spec.Template.Spec.Containers, corev1.Container{
			Name:  "metrics",
			Image: metricsImage,
			Env: []corev1.EnvVar{
				{
					Name: "REDIS_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: secretKey,
						},
					},
				},
				{
					Name:  "REDIS_ADDR",
					Value: "redis://localhost:6379",
				},
			},
			Ports: []corev1.ContainerPort{
				{
					Name:          "metrics",
					ContainerPort: 9121,
				},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("16Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
			},
		})
	}

	_ = ctrl.SetControllerReference(k, sts, scheme)
	return sts
}

// getContainerResources returns resource requirements for the container
func getContainerResources(k *keydbv1.Keydb) corev1.ResourceRequirements {
	// If resources are specified in spec, use them
	if !isEmptyResources(k.Spec.Resources) {
		return k.Spec.Resources
	}

	// Default resources if not specified
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}
}

// isEmptyResources checks if resource requirements are empty
func isEmptyResources(res corev1.ResourceRequirements) bool {
	return len(res.Requests) == 0 && len(res.Limits) == 0
}

// getPodAffinity returns pod affinity configuration for better pod distribution
func getPodAffinity(labels map[string]string) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}
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
