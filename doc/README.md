# secretgen-controller Documentation

`secretgen-controller` is a Kubernetes controller that generates and manages secrets declaratively. It provides Custom Resource Definitions (CRDs) to:

- **Generate** cryptographic material (certificates, RSA keys, SSH keys, and passwords) on-cluster
- **Share** secrets across namespaces via `SecretExport` and `SecretImport`
- **Compose** secrets from data residing in other Kubernetes resources via `SecretTemplate`

---

## Table of Contents

- [Install](install.md) — deploy secretgen-controller to your cluster
- [Quick Start](quickstart.md) — step-by-step walkthrough to generate your first secret
- [Concepts](concepts.md) — understand how the controller works

### Secret Generators

| CRD | API Group | Description |
|-----|-----------|-------------|
| [Password](password.md) | `secretgen.k14s.io/v1alpha1` | Generates random alpha-numeric passwords |
| [Certificate](certificate.md) | `secretgen.k14s.io/v1alpha1` | Generates X.509 certificates (root CA, intermediate CA, leaf) |
| [RSAKey](rsa-key.md) | `secretgen.k14s.io/v1alpha1` | Generates RSA key pairs |
| [SSHKey](ssh-key.md) | `secretgen.k14s.io/v1alpha1` | Generates SSH key pairs |

### Secret Sharing & Composition

| CRD | API Group | Description |
|-----|-----------|-------------|
| [SecretExport / SecretImport](secret-sharing.md) | `secretgen.carvel.dev/v1alpha1` | Export secrets from one namespace and import them into others |
| [SecretTemplate](secret-template.md) | `secretgen.carvel.dev/v1alpha1` | Compose a new Secret from data on other Kubernetes resources |

---

## API Groups

secretgen-controller uses two API groups:

| API Group | Purpose |
|-----------|---------|
| `secretgen.k14s.io/v1alpha1` | Secret generators (Password, Certificate, RSAKey, SSHKey) |
| `secretgen.carvel.dev/v1alpha1` | Secret sharing (SecretExport, SecretImport) and composition (SecretTemplate) |

> **Note:** `secretgen.carvel.dev` is the newer API group introduced in v0.5.0. Resources that previously existed under `secretgen.k14s.io` for sharing (SecretRequest) have been replaced by the resources in the new group.

---

## Examples Directory

Ready-to-use YAML files are provided in [`examples/`](../examples/):

| File | Description |
|------|-------------|
| `examples/passwords.yml` | Multiple password generation patterns |
| `examples/certs.yml` | CA chain + leaf certificates |
| `examples/rsa-key.yml` | RSA key pair |
| `examples/ssh-key.yml` | SSH key pair |
| `examples/secret-export.yml` | Secret sharing across namespaces |
| `examples/secret-export-image-pull-secret.yml` | Sharing image pull secrets |
| `examples/secret-template.yml` | Composing a secret from two input secrets |
| `examples/secret-template-non-secret-inputs.yml` | Composing from a Service and ConfigMap |
| `examples/certs-rotation/` | Certificate rotation with kapp |
