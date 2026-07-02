# Quick Start

This guide walks you through generating your first secrets, exporting them across namespaces, and composing secrets from other Kubernetes resources — all in a few minutes.

## Prerequisites

- secretgen-controller [installed](install.md) on your cluster
- `kubectl` configured

---

## Step 1 — Generate a Password

Apply the following manifest:

```yaml
# password-demo.yml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: my-password
  namespace: default
spec:
  length: 32
```

```bash
kubectl apply -f password-demo.yml
```

Check the `Password` resource status:

```bash
kubectl get password my-password
# NAME          DESCRIPTION           AGE
# my-password   Reconcile succeeded   5s
```

The controller creates a matching `Secret` with the same name:

```bash
kubectl get secret my-password -o jsonpath='{.data.password}' | base64 -d
# outputs a 32-character random string
```

---

## Step 2 — Generate a TLS Certificate Chain

```yaml
# certs-demo.yml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: my-root-ca
spec:
  isCA: true
  commonName: my-root-ca
---
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: my-app-cert
spec:
  caRef:
    name: my-root-ca
  alternativeNames:
  - my-app.default.svc.cluster.local
  extendedKeyUsage:
  - server_auth
```

```bash
kubectl apply -f certs-demo.yml
kubectl get certificate
# NAME          DESCRIPTION           AGE
# my-root-ca    Reconcile succeeded   4s
# my-app-cert   Reconcile succeeded   4s
```

Each certificate generates a `Secret` with two keys: `crt.pem` (certificate) and `key.pem` (private key):

```bash
kubectl get secret my-app-cert -o jsonpath='{.data.crt\.pem}' | base64 -d | openssl x509 -noout -subject -issuer
# subject=CN = my-app-cert
# issuer=CN = my-root-ca
```

---

## Step 3 — Share a Secret Across Namespaces

Create two namespaces and share a secret from one to the other:

```yaml
# sharing-demo.yml
apiVersion: v1
kind: Namespace
metadata:
  name: producer
---
apiVersion: v1
kind: Namespace
metadata:
  name: consumer
---
# Generate a secret in the producer namespace
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: shared-password
  namespace: producer
---
# Offer the secret to the consumer namespace
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: shared-password
  namespace: producer
spec:
  toNamespace: consumer
---
# Accept the secret in the consumer namespace
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretImport
metadata:
  name: shared-password
  namespace: consumer
spec:
  fromNamespace: producer
```

```bash
kubectl apply -f sharing-demo.yml

# Verify the secret was copied into the consumer namespace
kubectl get secret shared-password -n consumer
```

---

## Step 4 — Compose a Secret from Other Resources

```yaml
# template-demo.yml
apiVersion: v1
kind: Secret
metadata:
  name: db-user
stringData:
  username: admin
---
apiVersion: v1
kind: Secret
metadata:
  name: db-pass
stringData:
  password: s3cr3t
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
      name: db-user
  - name: pass
    ref:
      apiVersion: v1
      kind: Secret
      name: db-pass
  template:
    type: Opaque
    stringData:
      username: $(.user.data.username)
    data:
      password: $(.pass.data.password)
```

```bash
kubectl apply -f template-demo.yml
kubectl get secret db-credentials -o yaml
```

---

## Clean Up

```bash
kubectl delete -f password-demo.yml
kubectl delete -f certs-demo.yml
kubectl delete -f sharing-demo.yml
kubectl delete -f template-demo.yml
kubectl delete ns producer consumer
```

---

## Next Steps

- [Concepts](concepts.md) — understand ownership, reconciliation, and security model
- [Password reference](password.md) — all spec fields and examples
- [Certificate reference](certificate.md) — CA chains, leaf certs, custom projections
- [Secret Sharing reference](secret-sharing.md) — wildcard exports, label selectors, image-pull secrets
- [SecretTemplate reference](secret-template.md) — dynamic input names, non-secret inputs, RBAC
