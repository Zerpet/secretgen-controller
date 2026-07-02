# Secret Sharing — SecretExport and SecretImport

secretgen-controller enables secure cross-namespace secret sharing through a two-party consent model: the owner of a secret **exports** it, and the consumer **imports** it. A copy of the secret is created in the destination namespace only when both sides agree.

**API:** `secretgen.carvel.dev/v1alpha1`  
**Kinds:** `SecretExport`, `SecretImport`

> **Note:** Prior to v0.5.0 these resources were in the `secretgen.k14s.io` API group (as `SecretRequest`). Use the `secretgen.carvel.dev` group for all new resources.

---

## How It Works

```
producer namespace          consumer namespace
─────────────────           ──────────────────
Secret: db-password   ──►   Secret: db-password  (copy created by controller)
SecretExport          ◄──►  SecretImport
 toNamespace: consumer       fromNamespace: producer
```

1. A `Secret` exists in the **producer** namespace.
2. The producer creates a `SecretExport` (same name as the secret) pointing to the destination.
3. The consumer creates a `SecretImport` (same name) pointing back to the source.
4. The controller copies the secret into the consumer namespace and keeps it in sync.

---

## SecretExport Spec Reference

`SecretExport` must have the **same name** as the `Secret` it exports.

| Field | Type | Description |
|-------|------|-------------|
| `toNamespace` | `string` | Single destination namespace. Use `*` to export to **all** namespaces. |
| `toNamespaces` | `[]string` | List of destination namespaces. |
| `dangerousToNamespacesSelector` | `[]SelectorMatchField` | Export to namespaces matching label/annotation selectors. |

At least one of `toNamespace`, `toNamespaces`, or `dangerousToNamespacesSelector` must be set.

### SelectorMatchField

| Field | Type | Description |
|-------|------|-------------|
| `key` | `string` | JSONPath expression targeting a namespace field (e.g. `metadata.labels['env']`) |
| `operator` | `string` | `In`, `NotIn`, `Exists`, or `DoesNotExist` |
| `values` | `[]string` | Values to match (required for `In`/`NotIn`; must be omitted for `Exists`/`DoesNotExist`) |

---

## SecretImport Spec Reference

`SecretImport` must have the **same name** as the `SecretExport` (and the source `Secret`).

| Field | Type | Description |
|-------|------|-------------|
| `fromNamespace` | `string` | Source namespace; must match one of the `SecretExport`'s destination namespaces. |

---

## Examples

### Basic Export to a Single Namespace

```yaml
# In the 'platform' namespace
apiVersion: v1
kind: Secret
metadata:
  name: registry-credentials
  namespace: platform
type: kubernetes.io/dockerconfigjson
stringData:
  .dockerconfigjson: |
    {"auths":{"registry.example.com":{"username":"robot","password":"s3cr3t"}}}
---
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: registry-credentials
  namespace: platform
spec:
  toNamespace: app-team
---
# In the 'app-team' namespace
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretImport
metadata:
  name: registry-credentials
  namespace: app-team
spec:
  fromNamespace: platform
```

---

### Export to Multiple Namespaces

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: shared-api-key
  namespace: secrets-store
spec:
  toNamespaces:
  - team-alpha
  - team-beta
  - team-gamma
```

Each consuming namespace creates its own `SecretImport`:

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretImport
metadata:
  name: shared-api-key
  namespace: team-alpha
spec:
  fromNamespace: secrets-store
```

---

### Export to All Namespaces (Wildcard)

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: global-ca-cert
  namespace: pki
spec:
  toNamespace: "*"
```

Any namespace can then import it:

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretImport
metadata:
  name: global-ca-cert
  namespace: my-app
spec:
  fromNamespace: pki
```

**Excluding a Namespace from Wildcard Exports**

To prevent a namespace from receiving wildcard (`*`) exports, annotate it:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: restricted-ns
  annotations:
    secretgen.carvel.dev/excluded-from-wildcard-matching: ""
```

Secrets are only copied into `restricted-ns` when a `SecretExport` explicitly lists it in `toNamespace` or `toNamespaces`.

---

### Export Based on Namespace Labels

Export to all namespaces that belong to a specific project (e.g. using Rancher's project annotations):

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: project-db-password
  namespace: shared
spec:
  dangerousToNamespacesSelector:
  - key: "metadata.annotations['field.cattle.io/projectId']"
    operator: In
    values:
    - "cluster1:project-frontend"
```

---

### Combined: Static Namespaces + Label Selector

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: monitoring-credentials
  namespace: observability
spec:
  toNamespaces:
  - logging
  dangerousToNamespacesSelector:
  - key: "metadata.labels['team']"
    operator: In
    values:
    - platform
    - sre
```

---

### Generate and Share in One Manifest

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: secrets-store
---
apiVersion: v1
kind: Namespace
metadata:
  name: app-ns
---
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: db-password
  namespace: secrets-store
spec:
  length: 32
---
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: db-password
  namespace: secrets-store
spec:
  toNamespace: app-ns
---
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretImport
metadata:
  name: db-password
  namespace: app-ns
spec:
  fromNamespace: secrets-store
```

---

## Placeholder Secrets (Image Pull Secrets)

A **placeholder secret** is a regular `kubernetes.io/dockerconfigjson` secret annotated with `secretgen.carvel.dev/image-pull-secret: ""`. The controller automatically fills it with the **merged** `.dockerconfigjson` from all image-pull secrets exported to that namespace.

This pattern allows configuration (Helm charts, kapp-controller packages) to reference a single image pull secret without depending on secretgen-controller at authoring time.

### Example — Multiple Registry Credentials Merged

```yaml
# Global credentials exported to all namespaces
apiVersion: v1
kind: Secret
metadata:
  name: gcr-creds
  namespace: infra
type: kubernetes.io/dockerconfigjson
stringData:
  .dockerconfigjson: |
    {"auths":{"gcr.io":{"username":"_json_key","password":"..."}}}
---
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: gcr-creds
  namespace: infra
spec:
  toNamespace: "*"
---
# Private registry exported to a specific namespace only
apiVersion: v1
kind: Secret
metadata:
  name: private-reg-creds
  namespace: infra
type: kubernetes.io/dockerconfigjson
stringData:
  .dockerconfigjson: |
    {"auths":{"private.registry.io":{"username":"robot","password":"..."}}}
---
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: private-reg-creds
  namespace: infra
spec:
  toNamespace: prod
---
# Placeholder in the consuming namespace — filled automatically by the controller
apiVersion: v1
kind: Secret
metadata:
  name: combined-pull-secret
  namespace: prod
  annotations:
    secretgen.carvel.dev/image-pull-secret: ""
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: e30K  # empty JSON object — will be replaced
```

The `combined-pull-secret` in `prod` will contain merged credentials for both `gcr.io` and `private.registry.io`.

> **Security Warning:** Registry credentials exported to a namespace are readable by all users and workloads in that namespace. Export only credentials with the minimum required (read-only) scope.

---

## Status and Troubleshooting

Check the status of export/import resources:

```bash
kubectl get secretexport -n platform
# NAME                   DESCRIPTION           AGE
# registry-credentials   Reconcile succeeded   2m

kubectl get secretimport -n app-team
# NAME                   DESCRIPTION           AGE
# registry-credentials   Reconcile succeeded   2m
```

If a `SecretImport` shows an error, verify:
1. The `SecretExport` exists in the source namespace with a matching name.
2. The `SecretExport` lists the consumer namespace in `toNamespace`/`toNamespaces` or a matching selector.
3. The source `Secret` exists in the same namespace as the `SecretExport`.

---

## See Also

- [Concepts — Secret Sharing Security Model](concepts.md#secret-sharing-security-model)
- [Concepts — Placeholder Secrets](concepts.md#placeholder-secrets-image-pull-secrets)
