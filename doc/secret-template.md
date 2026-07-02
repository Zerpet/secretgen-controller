# SecretTemplate

`SecretTemplate` composes a new Kubernetes `Secret` from data residing in other Kubernetes resources (Secrets, ConfigMaps, Services, custom resources, etc.). It continuously watches the input resources and updates the composed secret whenever they change.

**API:** `secretgen.carvel.dev/v1alpha1`  
**Kind:** `SecretTemplate`

> **Note:** Available since v0.9.0 in the `secretgen.carvel.dev` API group.

---

## How It Works

```
Input Resource 1 (e.g. Secret)  ─┐
Input Resource 2 (e.g. Service) ─┼─► SecretTemplate ──► composed Secret
Input Resource N (e.g. Pod)     ─┘
```

1. Define one or more **input resources** — any Kubernetes API object the controller can read.
2. Define a **template** that extracts values from those resources using [JSONPath](https://kubernetes.io/docs/reference/kubectl/jsonpath/) expressions in `$(...)` syntax.
3. The controller creates (and keeps up to date) a `Secret` with the same name as the `SecretTemplate`.

---

## Spec Reference

| Field | Type | Description |
|-------|------|-------------|
| `inputResources` | `[]InputResource` | Ordered list of resources to read. Later entries can reference earlier ones dynamically. |
| `template` | object | Template for the output secret (see below). |
| `serviceAccountName` | `string` | Name of the `ServiceAccount` used to read input resources. Required when reading non-`Secret` resources. If omitted, only `Secret` resources can be used as inputs. |

### InputResource

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Identifier used to reference this resource in the template (e.g. `mySecret`). |
| `ref.apiVersion` | `string` | API version of the input resource (e.g. `v1`, `apps/v1`). |
| `ref.kind` | `string` | Kind of the input resource (e.g. `Secret`, `ConfigMap`, `Service`). |
| `ref.name` | `string` | Name of the input resource. May contain a `$(...)` JSONPath expression that references a previously resolved input resource for **dynamic name resolution**. |

### Template

| Field | Type | Description |
|-------|------|-------------|
| `type` | `string` | Kubernetes `Secret` type (immutable after first creation) |
| `stringData` | `map[string]string` | Plain-text key/value pairs for the output secret. Values support `$(...)` JSONPath expressions. |
| `data` | `map[string]string` | Base64-encoded key/value pairs. Use this to carry through already-base64-encoded data from a source `Secret`'s `.data` field. Values must be `$(...)` JSONPath expressions pointing to base64-encoded data. |
| `metadata.annotations` | `map[string]string` | Annotations for the output secret. Values support `$(...)`. |
| `metadata.labels` | `map[string]string` | Labels for the output secret. Values support `$(...)`. |

### JSONPath Expression Syntax

Expressions are wrapped in `$(` and `)`. The root of each expression is the name of an input resource:

| Example | Description |
|---------|-------------|
| `$(.mySecret.data.password)` | Read the `password` key from the `data` field of the `mySecret` input resource |
| `$(.mySecret.data.my\.key)` | Escape a `.` in the key name |
| `$(.service.spec.clusterIP)` | Read the cluster IP of a Service |
| `$(.service.spec.ports[?(@.name=="http")].port)` | Filter a port by name |
| `combined-$(.cfg.data.host)-$(.cfg.data.port)` | String interpolation with static text |

secretgen-controller uses the [Kubernetes JSONPath library](https://github.com/kubernetes/client-go/tree/master/util/jsonpath). Full documentation: [kubernetes.io/docs/reference/kubectl/jsonpath/](https://kubernetes.io/docs/reference/kubectl/jsonpath/).

---

## Examples

### Combine Two Secrets into One

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-username
stringData:
  username: admin
---
apiVersion: v1
kind: Secret
metadata:
  name: db-password
data:
  password: dG9wU2VjcmV0  # base64("topSecret")
---
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretTemplate
metadata:
  name: db-credentials
spec:
  inputResources:
  - name: user
    ref:
      apiVersion: v1
      kind: Secret
      name: db-username
  - name: pass
    ref:
      apiVersion: v1
      kind: Secret
      name: db-password
  template:
    type: Opaque
    stringData:
      username: $(.user.data.username)   # decoded automatically from data
    data:
      password: $(.pass.data.password)   # carried through as base64
```

---

### Read from ConfigMap and Service (Requires ServiceAccount)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  db-name: mydb
  db-host: postgres.default.svc.cluster.local
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  ports:
  - name: postgres
    port: 5432
---
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretTemplate
metadata:
  name: app-db-binding
spec:
  serviceAccountName: secret-reader
  inputResources:
  - name: config
    ref:
      apiVersion: v1
      kind: ConfigMap
      name: app-config
  - name: svc
    ref:
      apiVersion: v1
      kind: Service
      name: postgres
  template:
    type: Opaque
    stringData:
      host: $(.svc.spec.clusterIP)
      port: $(.svc.spec.ports[?(@.name=="postgres")].port)
      database: $(.config.data.db-name)
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: secret-reader
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
rules:
- apiGroups: [""]
  resources: [configmaps, services]
  verbs: [get, list, watch]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: secret-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: secret-reader
subjects:
- kind: ServiceAccount
  name: secret-reader
```

---

### Dynamic Input Name Resolution

The name of a later input resource can be determined by data from an earlier one. This is useful when the secret name is stored inside another resource (e.g. a pod's environment variable reference):

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretTemplate
metadata:
  name: helm-db-binding
spec:
  serviceAccountName: helm-reader
  inputResources:
  - name: pod
    ref:
      apiVersion: v1
      kind: Pod
      name: my-release-postgresql-0
  - name: db-secret
    ref:
      apiVersion: v1
      kind: Secret
      # The secret name is read from the pod's environment variable reference
      name: $(.pod.spec.containers[?(@.name=="postgresql")].env[?(@.name=="POSTGRES_PASSWORD")].valueFrom.secretKeyRef.name)
  template:
    type: Opaque
    data:
      password: $(.db-secret.data.postgres-password)
```

---

### Service Binding for a Helm-Installed PostgreSQL

This example reads resources created by `helm install my-release bitnami/postgresql` and produces a [Service Binding](https://github.com/servicebinding/spec)-compatible secret:

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretTemplate
metadata:
  name: postgres-binding
spec:
  serviceAccountName: helm-reader
  inputResources:
  - name: pod
    ref:
      apiVersion: v1
      kind: Pod
      name: my-release-postgresql-0
  - name: service
    ref:
      apiVersion: v1
      kind: Service
      name: my-release-postgresql
  - name: secret
    ref:
      apiVersion: v1
      kind: Secret
      name: $(.pod.spec.containers[?(@.name=="postgresql")].env[?(@.name=="POSTGRES_PASSWORD")].valueFrom.secretKeyRef.name)
  template:
    metadata:
      labels:
        app: my-release-postgresql
    type: postgresql
    stringData:
      host: $(.service.spec.clusterIP)
      port: $(.service.spec.ports[?(@.name=="tcp-postgresql")].port)
      database: postgres
      username: postgres
    data:
      password: $(.secret.data.postgres-password)
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: helm-reader
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: helm-reader
rules:
- apiGroups: [""]
  resources: [secrets, services, pods]
  verbs: [get, list, watch]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: helm-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: helm-reader
subjects:
- kind: ServiceAccount
  name: helm-reader
```

---

### Add Labels and Annotations to the Output Secret

```yaml
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretTemplate
metadata:
  name: annotated-secret
spec:
  inputResources:
  - name: source
    ref:
      apiVersion: v1
      kind: Secret
      name: raw-credentials
  template:
    metadata:
      labels:
        app.kubernetes.io/managed-by: secretgen-controller
        environment: production
      annotations:
        description: Credentials managed by secretgen-controller
    type: Opaque
    data:
      password: $(.source.data.password)
```

---

## RBAC Requirements

When `serviceAccountName` is set, the referenced `ServiceAccount` must have `get`, `list`, and `watch` permissions for every resource kind used in `inputResources`. Create a `Role` and `RoleBinding` in the **same namespace** as the `SecretTemplate`.

For cluster-scoped resources, use `ClusterRole` and `ClusterRoleBinding`.

Minimal RBAC template (fill in resource kinds):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: <name>
rules:
- apiGroups: [""]
  resources: [<resource-kinds>]
  verbs: [get, list, watch]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: <name>
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: <name>
subjects:
- kind: ServiceAccount
  name: <service-account-name>
```

---

## Status and Troubleshooting

```bash
kubectl get secrettemplate my-template
# NAME          DESCRIPTION           AGE
# my-template   Reconcile succeeded   1m
```

Common issues:

| Symptom | Likely Cause |
|---------|-------------|
| `permission denied` error | ServiceAccount is missing RBAC permissions for an input resource |
| Input resource not found | The resource name, namespace, or API version is incorrect |
| Template not updating | The controller watches input resources; ensure you haven't removed a watch verb from the RBAC |
| `type` cannot be changed | The `Secret` type field is immutable; delete and re-create to change it |

---

## See Also

- [Concepts — Secret Templates field](concepts.md#secret-templates-secrettemplate-field)
- [Secret Sharing](secret-sharing.md) — export the composed secret to other namespaces
