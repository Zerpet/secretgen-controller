# Concepts

## How secretgen-controller Works

secretgen-controller is a standard Kubernetes controller that watches custom resources and reconciles them towards a desired state. When you create a `Password`, `Certificate`, `RSAKey`, `SSHKey`, or `SecretTemplate` resource, the controller generates or assembles the required secret material and stores it in a regular Kubernetes `Secret` with the same name in the same namespace.

---

## Secret Ownership

By default, generated `Secret` resources are owned by their corresponding custom resource (e.g. a `Password`) via `metadata.ownerReferences`. This means:

- Deleting the custom resource also deletes the generated `Secret`.
- If you want to **keep** the `Secret` after deleting the custom resource, remove its `ownerReferences`:

```bash
kubectl patch secret my-password \
  -p '{"metadata":{"ownerReferences":null}}'
```

---

## Reconciliation & Idempotency

The controller continuously watches for changes. It will:

- Re-generate a secret if the owning custom resource is updated (e.g. you change the `length` of a `Password`).
- Re-copy a shared secret when the source changes (for `SecretExport`/`SecretImport`).
- Re-compose a `SecretTemplate` output when any input resource changes.

> **Note:** Generated cryptographic material (certificates, keys) is **not** rotated automatically unless you explicitly delete and re-create the custom resource, or trigger a change to its spec. See the [`examples/certs-rotation/`](../examples/certs-rotation/) directory for a kapp-based rotation pattern.

---

## API Groups

| API Group | Resources |
|-----------|-----------|
| `secretgen.k14s.io/v1alpha1` | `Password`, `Certificate`, `RSAKey`, `SSHKey` |
| `secretgen.carvel.dev/v1alpha1` | `SecretExport`, `SecretImport`, `SecretTemplate` |

The `secretgen.carvel.dev` group was introduced in v0.5.0 and is the current canonical group for sharing and composition resources.

---

## Secret Templates (`secretTemplate` field)

Every generator CRD (`Password`, `Certificate`, `RSAKey`, `SSHKey`) supports an optional `secretTemplate` field that lets you control the shape of the generated `Secret`:

| Sub-field | Type | Description |
|-----------|------|-------------|
| `type` | `string` | Overrides the Kubernetes `Secret` type (default varies per generator) |
| `stringData` | `map[string]string` | Overrides the keys and values written to `Secret.data`. Values support variable interpolation with `$(varName)` syntax. |

Each generator exposes a different set of variables for interpolation; see the individual reference pages for details.

**Example** — project a `Password` into an Opaque secret under a custom key:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: pg-password
spec:
  length: 40
  secretTemplate:
    type: Opaque
    stringData:
      postgresql-password: $(value)
```

---

## Secret Sharing Security Model

Kubernetes scopes `Secret` access to a namespace — a pod in `namespace-a` cannot reference a `Secret` in `namespace-b`. secretgen-controller provides secure cross-namespace sharing via a two-party consent model:

1. **The exporter** (owner of the source `Secret`) creates a `SecretExport` in the source namespace, explicitly naming the destination namespace(s).
2. **The importer** (user in the destination namespace) creates a `SecretImport` in the destination namespace, explicitly naming the source namespace.

A secret is copied **only** when both resources agree. Neither side alone can initiate the copy, preventing accidental or malicious exfiltration.

---

## Placeholder Secrets (Image Pull Secrets)

As an alternative to `SecretImport`, a plain Kubernetes `Secret` of type `kubernetes.io/dockerconfigjson` annotated with `secretgen.carvel.dev/image-pull-secret: ""` acts as a **placeholder**. The controller automatically fills it with all image-pull secrets that have been exported to the containing namespace.

This allows Helm charts, kapp-controller packages, and other configuration to reference image pull secrets without taking a hard dependency on secretgen-controller — the plain `Secret` resource works in any environment.

> **Warning:** Exporting registry credentials makes them visible to all users in the destination namespace(s). Export only credentials with the minimum necessary (read-only) scope.

---

## Status Conditions

All CRDs expose a `.status.friendlyDescription` field printed by `kubectl get`. A successful reconciliation shows `Reconcile succeeded`. If an error occurs, the description contains a human-readable message.

```bash
kubectl get password my-password
# NAME          DESCRIPTION           AGE
# my-password   Reconcile succeeded   1m
```
