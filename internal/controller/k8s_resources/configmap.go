package k8sresources

import (
	"fmt"
	"strings"

	keydbv1 "github.com/rsingh0101/keydb-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GenerateKeydbConfigMap(k *keydbv1.Keydb, scheme *runtime.Scheme) ([]*corev1.ConfigMap, error) {
	labels := map[string]string{"apps": k.Name}

	// Base config
	config := []string{
		"bind 0.0.0.0 ::",
		"protected-mode no",
		"dir /bitnami/keydb/data",
		"port 6379",
		"loglevel notice",
		"appendonly yes",
		fmt.Sprintf("requirepass %s", k.Spec.Password),
	}

	// Only add replication bits if enabled
	if k.Spec.Replication.Enabled {
		switch k.Spec.Replication.Mode {
		case "master-replica":
			host := NormalizeFQDN(k.Spec.Replication.Domain[0])
			port := k.Spec.Replication.Port
			config = append(config,
				fmt.Sprintf("replicaof %s %d", host, port),
				fmt.Sprintf("masterauth %s", k.Spec.Password),
				"replica-read-consistency yes",
				"repl-diskless-sync yes",
				"repl-diskless-sync-delay 0",
			)
		case "master-master":
			// Active-active multi-master mesh
			config = append(config,
				"active-replica yes",
				"multi-master yes",
				"replica-read-only no",
				"repl-diskless-sync yes",
				"repl-diskless-sync-delay 0",
			)

			// Generate replicaof lines for all pods in the StatefulSet
			for _, domain := range k.Spec.Replication.Domain {
				host := NormalizeFQDN(domain)
				config = append(config,
					fmt.Sprintf("replicaof %s %d", host, k.Spec.Replication.Port),
					fmt.Sprintf("masterauth %s", k.Spec.Password),
				)
			}
		}
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name + "-config",
			Namespace: k.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"keydb.conf": strings.Join(config, "\n") + "\n",
		},
	}

	if err := ctrl.SetControllerReference(k, cm, scheme); err != nil {
		return nil, err
	}

	health_cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name + "-healthz",
			Namespace: k.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"ping_readiness_local.sh": `#!/bin/bash
				. /opt/bitnami/scripts/keydb-env.sh
				. /opt/bitnami/scripts/liblog.sh
				response=$(
					timeout -s 15 $1 \
					keydb-cli \
						-h localhost \
						-a "$KEYDB_PASSWORD" \
						-p $KEYDB_PORT_NUMBER \
						ping
				)
				if [[ "$?" -eq "124" ]]; then
					error "Timed out"
					exit 1
				fi
				if [[ "$response" != "PONG" ]]; then
					error "$response"
					exit 1
				fi`,
			"ping_liveness_local.sh": `#!/bin/bash

			. /opt/bitnami/scripts/keydb-env.sh
			. /opt/bitnami/scripts/liblog.sh

			response=$(
				timeout -s 15 $1 \
				keydb-cli \
					-h localhost \
					-a "$KEYDB_PASSWORD" \
					-p $KEYDB_PORT_NUMBER \
					ping
			)
			if [[ "$?" -eq "124" ]]; then
				error "Timed out"
				exit 1
			fi
			responseFirstWord="$(echo "$response" | head -n1 | awk '{print $1;}')"
			if [[ "$response" != "PONG" ]] && [[ "$responseFirstWord" != "LOADING" ]] && [[ "$responseFirstWord" != "MASTERDOWN" ]]; then
				error "$response"
				exit 1
			fi`,
			"ping_readiness_master.sh": `#!/bin/bash

			. /opt/bitnami/scripts/keydb-env.sh
			. /opt/bitnami/scripts/liblog.sh

			response=$(
				timeout -s 15 $1 \
				keydb-cli \
					-h keydb-master \
					-p 6379 \
					-a "$KEYDB_MASTER_PASSWORD" \
					ping
			)
			if [[ "$?" -eq "124" ]]; then
				error "Timed out"
				exit 1
			fi
			if [[ "$response" != "PONG" ]]; then
				error "$response"
				exit 1
			fi`,
			"ping_liveness_master.sh": `#!/bin/bash
					. /opt/bitnami/scripts/keydb-env.sh
					. /opt/bitnami/scripts/liblog.sh
					response=$(
						timeout -s 15 $1 \
						keydb-cli \
							-h keydb-master \
							-p 6379 \
							-a "$KEYDB_MASTER_PASSWORD" \
							ping
					)
					if [[ "$?" -eq "124" ]]; then
						error "Timed out"
						exit 1
					fi
					responseFirstWord="$(echo "$response" | head -n1 | awk '{print $1;}')"
					if [[ "$response" != "PONG" ]] && [[ "$responseFirstWord" != "LOADING" ]]; then
						error "$response"
						exit 1
					fi`,
			"ping_readiness_local_and_master.sh": `#!/bin/bash
			 #!/bin/bash

				script_dir="$(dirname "$0")"
				exit_status=0
				"$script_dir/ping_readiness_local.sh" $1 || exit_status=$?
				"$script_dir/ping_readiness_master.sh" $1 || exit_status=$?
				exit $exit_status`,
			"ping_liveness_local_and_master.sh": `#!/bin/bash

				script_dir="$(dirname "$0")"
				exit_status=0
				"$script_dir/ping_liveness_local.sh" $1 || exit_status=$?
				"$script_dir/ping_liveness_master.sh" $1 || exit_status=$?
				exit $exit_status
			`,
		},
	}

	return []*corev1.ConfigMap{cm, health_cm}, nil

}
