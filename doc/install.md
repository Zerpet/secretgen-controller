# Install secretgen-controller

## Prerequisites

- A running Kubernetes cluster (v1.20+)
- `kubectl` configured to talk to the cluster
- Cluster-admin permissions (required for CRD installation)

---

## Quick Install

Grab the latest release YAML and apply it with `kubectl` or [kapp](https://carvel.dev/kapp/):

```bash
# Using kubectl
kubectl apply -f https://github.com/carvel-dev/secretgen-controller/releases/latest/download/release.yml

# Using kapp (recommended — tracks all deployed resources)
kapp deploy -a secretgen-controller \
  -f https://github.com/carvel-dev/secretgen-controller/releases/latest/download/release.yml
```

This installs:
- All CRDs (`Certificate`, `Password`, `RSAKey`, `SSHKey`, `SecretExport`, `SecretImport`, `SecretTemplate`)
- The controller `Deployment` in the `secretgen-controller-system` namespace
- The necessary `ServiceAccount`, `ClusterRole`, and `ClusterRoleBinding`

### Verify the Installation

```bash
kubectl get pods -n secretgen-controller-system
```

Expected output:
```
NAME                                      READY   STATUS    RESTARTS   AGE
secretgen-controller-xxxx-yyyy            1/1     Running   0          30s
```

Check that the CRDs are registered:

```bash
kubectl get crds | grep -E 'secretgen|carvel'
```

Expected output (abbreviated):
```
certificates.secretgen.k14s.io
passwords.secretgen.k14s.io
rsakeys.secretgen.k14s.io
sshkeys.secretgen.k14s.io
secretexports.secretgen.carvel.dev
secretimports.secretgen.carvel.dev
secrettemplates.secretgen.carvel.dev
```

---

## Install a Specific Version

Find a version on the [Releases page](https://github.com/carvel-dev/secretgen-controller/releases) and replace `<version>` below:

```bash
kubectl apply -f https://github.com/carvel-dev/secretgen-controller/releases/download/<version>/release.yml
```

---

## Advanced: Build from Source

`release.yml` is produced with [ytt](https://carvel.dev/ytt/) and [kbld](https://carvel.dev/kbld/). You can customise the configuration by building from source:

```bash
git clone https://github.com/carvel-dev/secretgen-controller.git
cd secretgen-controller

# Deploy using ytt + kbld + kapp (add -v image_repo=... to push to your registry)
kapp deploy -a secretgen-controller -f <(ytt -f config/ | kbld -f -)
```

---

## Uninstall

```bash
# Using kubectl
kubectl delete -f https://github.com/carvel-dev/secretgen-controller/releases/latest/download/release.yml

# Using kapp
kapp delete -a secretgen-controller
```

> **Warning:** Deleting the CRDs will also delete all custom resources of those types. Make sure any generated secrets you want to keep have their `metadata.ownerReferences` cleared beforehand, or take a backup.

---

Next: [Quick Start →](quickstart.md)
