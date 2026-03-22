# KeyDB Kubernetes Operator

<img src="https://github.com/rsingh0101/keydb-operator/blob/main/images/k8skeydb.png" width="300" height="300">

A production-ready Kubernetes operator for managing KeyDB clusters with advanced scaling, high availability, and observability features.

## Description

The KeyDB Operator simplifies the deployment and management of KeyDB (a high-performance fork of Redis) clusters on Kubernetes. It provides comprehensive lifecycle management, automatic scaling, high availability, and production-grade features.

### Key Features

**Core Functionality:**
- **Lifecycle Management** - Automated creation, updates, and deletion of KeyDB clusters
- **Replication** - Support for master-replica and master-master replication modes
- **Persistence** - Configurable persistent storage with PVC support
- **Auto-healing** - Automatic pod recovery and health monitoring

**Production Features:**
- **Resource Management** - CPU and memory requests/limits with sensible defaults
- **Pod Anti-Affinity** - Automatic pod distribution across nodes for fault tolerance
- **Pod Disruption Budget** - Protects against accidental downtime during updates
- **Enhanced Status** - Comprehensive status tracking with conditions and replica health
- **Health Checks** - Liveness and readiness probes for reliable operation

**Scaling & Reliability:**
- **Horizontal Scaling** - Scale replicas up/down with validation
- **Fault Tolerance** - Quorum-based protection for master-master clusters
- **Status Observability** - Real-time status with conditions, replica tracking, and health metrics
- **Graceful Operations** - Safe scaling and update operations

For detailed analysis and future improvements, see [SCALING_ANALYSIS.md](SCALING_ANALYSIS.md).

## Quick Start

### Prerequisites
- Go version v1.24.0+
- Docker version 17.03+
- kubectl version v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster
- kubebuilder (for development)

### Installation

**1. Install the CRDs:**
```sh
make install
```

**2. Build and deploy the operator:**
```sh
# Build the image
make docker-build IMG=<your-registry>/keydb-operator:latest

# Push the image
make docker-push IMG=<your-registry>/keydb-operator:latest

# Deploy the operator
make deploy IMG=<your-registry>/keydb-operator:latest
```

**3. Verify the operator is running:**
```sh
kubectl get deployment -n keydb-operator-system
kubectl get pods -n keydb-operator-system
```

**4. Create a KeyDB instance:**
```sh
kubectl apply -f config/samples/keydb_v1_keydb.yaml
```

**5. Check the status:**
```sh
kubectl get keydb
kubectl describe keydb <keydb-name>
```

## Getting Started

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/keydb-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/keydb-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**

You can apply the samples (examples) from the config/samples:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples have default values to test it out.

## Usage Examples

### Basic KeyDB Instance

```yaml
apiVersion: keydb.keydb/v1
kind: Keydb
metadata:
  name: keydb-basic
  namespace: default
spec:
  replicas: 3
  image: docker.io/bitnamilegacy/keydb:6.3.4-debian-12-r24
  password: my-secure-password
  persistence:
    enabled: true
    size: 10Gi
    storageClassName: standard
  replication:
    mode: master-master
    enabled: true
    port: 6379
```

### Production-Ready Configuration

```yaml
apiVersion: keydb.keydb/v1
kind: Keydb
metadata:
  name: keydb-production
  namespace: production
spec:
  replicas: 5
  image: docker.io/bitnamilegacy/keydb:6.3.4-debian-12-r24
  password: secure-password
  
  # Resource management
  resources:
    requests:
      cpu: "500m"
      memory: "1Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
  
  # Persistence
  persistence:
    enabled: true
    size: 50Gi
    storageClassName: ssd
  
  # Replication
  replication:
    mode: master-master
    enabled: true
    port: 6379
```

### Checking Status

```sh
# Get basic status
kubectl get keydb

# Detailed status with conditions
kubectl get keydb <name> -o yaml

# Describe for human-readable output
kubectl describe keydb <name>
```

The status includes:
- **Phase**: Current phase (Pending, Scaling, Running, Unknown)
- **Conditions**: Ready, Progressing, Degraded conditions
- **Replicas**: Detailed replica status (ready, not ready, failed)
- **ReadyReplicas**: Count of ready replicas
- **CurrentReplicas**: Current replica count

### Monitoring

```sh
# Check Pod Disruption Budget
kubectl get pdb

# Verify pod distribution (anti-affinity)
kubectl get pods -o wide

# Check resource usage
kubectl top pods -l apps=<keydb-name>
```

## API Reference

### KeydbSpec

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `image` | string | KeyDB container image | Yes |
| `replicas` | int32 | Number of replicas (min: 1) | Yes |
| `password` | string | KeyDB password | Yes |
| `replication` | ReplicationSpec | Replication configuration | No |
| `persistence` | PersistenceSpec | Persistence configuration | No |
| `resources` | ResourceRequirements | CPU/memory requests and limits | No |

### ReplicationSpec

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | bool | Enable replication |
| `mode` | string | Replication mode: "master-replica" or "master-master" |
| `domain` | []string | External KeyDB domains for replication |
| `port` | int32 | Replication port (default: 6379) |

### PersistenceSpec

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | bool | Enable persistent storage |
| `size` | string | Storage size (e.g., "10Gi") |
| `storageClassName` | string | Storage class name |

### KeydbStatus

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current phase (Pending, Scaling, Running, Unknown) |
| `conditions` | []Condition | Status conditions |
| `replicas` | ReplicaStatus | Per-replica status |
| `readyReplicas` | int32 | Number of ready replicas |
| `currentReplicas` | int32 | Current replica count |
| `observedGeneration` | int64 | Observed generation |

## Features in Detail

### Resource Management

The operator automatically applies resource requests and limits. If not specified, defaults are:
- **Requests**: 100m CPU, 128Mi memory
- **Limits**: 2 CPU, 2Gi memory

You can override these in the spec:
```yaml
spec:
  resources:
    requests:
      cpu: "500m"
      memory: "1Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
```

### Pod Anti-Affinity

Pods are automatically distributed across nodes using preferred pod anti-affinity. This ensures:
- Better fault tolerance
- Improved resource distribution
- Reduced single point of failure

### Pod Disruption Budget

A Pod Disruption Budget is automatically created to protect your cluster:
- **Master-Master**: Maintains quorum (majority of replicas)
- **Other modes**: Allows at most 1 pod disruption

### Enhanced Status

The operator provides comprehensive status information:
- **Conditions**: Kubernetes-standard conditions for Ready, Progressing states
- **Replica Tracking**: Lists of ready, not ready, and failed pods
- **Health Metrics**: Real-time replica counts and health status

## Troubleshooting

### Check Operator Logs

```sh
kubectl logs -n keydb-operator-system deployment/keydb-operator-controller-manager
```

### Check KeyDB Pod Logs

```sh
kubectl logs <keydb-pod-name>
```

### Common Issues

**Pods not starting:**
- Check resource availability: `kubectl describe pod <pod-name>`
- Verify storage class: `kubectl get storageclass`
- Check events: `kubectl get events --sort-by='.lastTimestamp'`

**Replication issues:**
- Verify network connectivity between pods
- Check KeyDB logs for replication errors
- Ensure correct domain configuration

**Status not updating:**
- Check operator logs for errors
- Verify RBAC permissions
- Ensure StatefulSet is created

## Development

### Building

```sh
# Generate code
make generate

# Generate manifests
make manifests

# Build binary
make build

# Run locally
make run
```

### Testing

```sh
# Run unit tests
make test

# Run e2e tests (requires Kind)
make test-e2e
```

### Code Generation

After modifying API types, regenerate code:
```sh
make generate
make manifests
```

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/keydb-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/keydb-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Architecture

The operator follows the Kubernetes operator pattern:
- **Controller**: Watches KeyDB custom resources and reconciles desired state
- **StatefulSet**: Manages KeyDB pods with stable network identities
- **Services**: Headless and ClusterIP services for pod discovery
- **ConfigMaps**: KeyDB configuration management
- **Secrets**: Secure password storage
- **PDB**: Pod Disruption Budget for availability protection

## Roadmap

See [SCALING_ANALYSIS.md](SCALING_ANALYSIS.md) for detailed analysis and future improvements.

**Planned Features:**
- Horizontal Pod Autoscaling (HPA)
- KeyDB metrics exporter integration
- Backup and restore functionality
- TLS/SSL support
- Advanced configuration options
- Graceful scaling with replication coordination

## Contributing

Contributions are welcome! Please see our contributing guidelines.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Documentation

- [Scaling Analysis](SCALING_ANALYSIS.md) - Comprehensive analysis of scaling improvements
- [Implementation Examples](IMPLEMENTATION_EXAMPLES.md) - Code examples for features
- [Implementation Summary](IMPLEMENTATION_SUMMARY.md) - Summary of implemented features
- [Scaling Checklist](SCALING_CHECKLIST.md) - Checklist for scaling improvements

## License

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

