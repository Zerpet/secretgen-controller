# Certificate

`Certificate` generates an X.509 certificate backed by a 3072-bit RSA key and stores both in a Kubernetes `Secret`.

**API:** `secretgen.k14s.io/v1alpha1`  
**Kind:** `Certificate`

---

## Default Secret

By default the generated `Secret` has:

| Field | Value |
|-------|-------|
| `type` | `Opaque` |
| `data.crt.pem` | PEM-encoded X.509 certificate |
| `data.key.pem` | PEM-encoded RSA private key |

---

## Spec Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isCA` | `bool` | `false` | Set to `true` for root or intermediate CA certificates. CA certs have key usage `KeyUsageCertSign` + `KeyUsageCRLSign`; leaf certs have `KeyUsageKeyEncipherment` + `KeyUsageDigitalSignature`. |
| `caRef` | `LocalObjectReference` | — | Name of the CA `Certificate` resource used to sign this certificate. Omit for self-signed root CAs. |
| `commonName` | `string` | — | Certificate CN field |
| `organization` | `string` | — | Certificate Organization field |
| `alternativeNames` | `[]string` | — | Subject Alternative Names — IP addresses or DNS names |
| `extendedKeyUsage` | `[]string` | — | Extended key usages. Supported values: `client_auth`, `server_auth` |
| `duration` | `int64` | `365` | Certificate validity in **days** from now |
| `secretTemplate` | object | — | Customise the generated `Secret` (see below) |

### `secretTemplate` Sub-fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | `string` | Kubernetes `Secret` type |
| `stringData` | `map[string]string` | Key/value pairs; values may reference the variables below |

**Available variables:**

| Variable | Description |
|----------|-------------|
| `$(certificate)` | PEM-encoded certificate |
| `$(privateKey)` | PEM-encoded private key |

---

## Examples

### Self-Signed Root CA

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: root-ca
spec:
  isCA: true
  commonName: my-root-ca
  organization: My Org
  duration: 3650  # 10 years
```

---

### Intermediate CA Signed by a Root CA

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: intermediate-ca
spec:
  isCA: true
  caRef:
    name: root-ca
  commonName: my-intermediate-ca
  duration: 1825  # 5 years
```

---

### Leaf Certificate for TLS

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: app-tls
spec:
  caRef:
    name: intermediate-ca
  commonName: my-app
  alternativeNames:
  - my-app.default.svc.cluster.local
  - my-app.example.com
  extendedKeyUsage:
  - server_auth
  duration: 90
```

---

### Mutual TLS — Server + Client Certificates

```yaml
# CA
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: mtls-ca
spec:
  isCA: true
  commonName: mtls-ca
---
# Server cert
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: server-cert
spec:
  caRef:
    name: mtls-ca
  alternativeNames:
  - server.default.svc.cluster.local
  extendedKeyUsage:
  - server_auth
---
# Client cert
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: client-cert
spec:
  caRef:
    name: mtls-ca
  extendedKeyUsage:
  - client_auth
```

---

### Custom Secret Projection

Store the certificate and key under application-specific key names:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: app-cert
spec:
  caRef:
    name: root-ca
  alternativeNames:
  - app.example.com
  secretTemplate:
    type: kubernetes.io/tls
    stringData:
      tls.crt: $(certificate)
      tls.key: $(privateKey)
```

---

### Three-Level Certificate Chain

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: root-ca
spec:
  isCA: true
---
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: inter-ca
spec:
  isCA: true
  caRef:
    name: root-ca
---
apiVersion: secretgen.k14s.io/v1alpha1
kind: Certificate
metadata:
  name: leaf-cert
spec:
  caRef:
    name: inter-ca
  alternativeNames:
  - app.internal
```

---

## Using the Generated Secret

Mount the certificate and key in a pod:

```yaml
volumes:
- name: tls
  secret:
    secretName: app-tls
containers:
- name: app
  volumeMounts:
  - name: tls
    mountPath: /etc/tls
    readOnly: true
```

The files at `/etc/tls/crt.pem` and `/etc/tls/key.pem` will contain the certificate and private key respectively (or your custom key names if you used `secretTemplate`).

---

## Certificate Rotation

Certificates are not rotated automatically. To rotate, delete the `Certificate` resource and re-create it (or modify its spec). The [`examples/certs-rotation/`](../examples/certs-rotation/) directory shows a kapp-based pattern for rotation that replaces pods consuming the rotated certificate.

---

## See Also

- [RSA Key](rsa-key.md) — generate a standalone RSA key pair
- [Secret Sharing](secret-sharing.md) — export the certificate secret to other namespaces
