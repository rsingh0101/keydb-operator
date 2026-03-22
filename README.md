# 🚀 KeyDB Kubernetes Operator

<p>
  <b>Production-ready Kubernetes Operator for KeyDB</b><br/>
  Automated lifecycle, replication, scaling, and resilience management.
</p>

![Kubernetes](https://img.shields.io/badge/kubernetes-1.11+-blue.svg)
![Go](https://img.shields.io/badge/go-1.24+-blue.svg)
[![CI/CD](https://github.com/rsingh0101/keydb-operator/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/rsingh0101/keydb-operator/actions)
![Lint](https://img.shields.io/badge/lint-golangci--lint-blue)

---

## 📌 Overview

The **KeyDB Operator** extends Kubernetes to manage **KeyDB clusters declaratively**, enabling automated provisioning, scaling, failover, and recovery.

Built using **Kubebuilder/controller-runtime**, it follows Kubernetes-native reconciliation patterns and is suitable for **production deployments**.

---

## ✨ Features

- 🔄 Lifecycle Management (Create / Update / Delete)
- 🔁 Replication (Master-Replica, Master-Master)
- 📈 Horizontal Scaling
- 🩺 Self-Healing (Pod recovery)
- 🌐 Flexible Topology
- ⚡ Fault Tolerance
- 📊 Prometheus Metrics Exposure
- 🔍 Observability via CR status & events

---

## 🏗️ Architecture

```
User (CR YAML)
     ↓
KeyDB Custom Resource (CRD)
     ↓
Operator (Controller)
     ↓
Kubernetes Resources (Pods, Services, PVCs)
     ↓
KeyDB Cluster
```

---

## ⚙️ Prerequisites

- Go ≥ 1.24
- Docker ≥ 17.03
- kubectl ≥ 1.11
- Kubernetes Cluster ≥ 1.11
- (Optional) Kind for local testing

---

## 🚀 Quick Start

### 1. Clone Repository
```
git clone https://github.com/rsingh0101/keydb-operator.git
cd keydb-operator
```

### 2. Build & Push Image
```
make docker-build docker-push IMG=<registry>/keydb-operator:tag
```

### 3. Install CRDs
```
make install
```

### 4. Deploy Operator
```
make deploy IMG=<registry>/keydb-operator:tag
```

### 5. Deploy Sample CR
```
kubectl apply -k config/samples/
```

### 6. Verify
```
kubectl get pods -A
kubectl get keydbs
```

---

## 🧪 Testing (E2E)

```
make test-e2e
```

---

## 🔄 CI/CD Pipeline

```
Lint → Build → Test → E2E → Docker Build → Deploy
```

---

## 📦 Deployment Options

### YAML Bundle
```
make build-installer IMG=<registry>/keydb-operator:tag
kubectl apply -f dist/install.yaml
```

---

### Helm Chart
```
kubebuilder edit --plugins=helm/v1-alpha
```

---

## 🧹 Uninstall

```
kubectl delete -k config/samples/
make uninstall
make undeploy
```
