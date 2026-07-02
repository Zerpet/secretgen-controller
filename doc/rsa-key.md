# RSAKey

`RSAKey` generates an RSA key pair and stores it in a Kubernetes `Secret`.

**API:** `secretgen.k14s.io/v1alpha1`  
**Kind:** `RSAKey`

---

## Default Secret

By default the generated `Secret` has:

| Field | Value |
|-------|-------|
| `type` | `Opaque` |
| `data.pub.pem` | PEM-encoded RSA public key |
| `data.key.pem` | PEM-encoded RSA private key |

---

## Spec Reference

| Field | Type | Description |
|-------|------|-------------|
| `secretTemplate` | object | Customise the generated `Secret` (see below) |

### `secretTemplate` Sub-fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | `string` | Kubernetes `Secret` type |
| `stringData` | `map[string]string` | Key/value pairs; values may reference the variables below |

**Available variables:**

| Variable | Description |
|----------|-------------|
| `$(publicKey)` | PEM-encoded RSA public key |
| `$(privateKey)` | PEM-encoded RSA private key |

---

## Examples

### Minimal — Default Key Pair

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: RSAKey
metadata:
  name: my-rsa-key
spec: {}
```

Produces a `Secret` named `my-rsa-key` with keys `pub.pem` and `key.pem`.

---

### Custom Secret Projection

Store the keys under application-specific names:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: RSAKey
metadata:
  name: jwt-signing-key
spec:
  secretTemplate:
    type: Opaque
    stringData:
      jwt-public-key: $(publicKey)
      jwt-private-key: $(privateKey)
```

---

### Private Key Only

If the consuming application only needs the private key:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: RSAKey
metadata:
  name: signing-key
spec:
  secretTemplate:
    type: Opaque
    stringData:
      private.pem: $(privateKey)
```

---

## Using the Generated Secret

Reference the private key in a deployment:

```yaml
env:
- name: RSA_PRIVATE_KEY
  valueFrom:
    secretKeyRef:
      name: jwt-signing-key
      key: jwt-private-key
```

Or mount as a file:

```yaml
volumes:
- name: rsa-keys
  secret:
    secretName: my-rsa-key
containers:
- name: app
  volumeMounts:
  - name: rsa-keys
    mountPath: /etc/rsa
    readOnly: true
# Files available at: /etc/rsa/pub.pem and /etc/rsa/key.pem
```

---

## See Also

- [SSHKey](ssh-key.md) — generate an SSH key pair
- [Certificate](certificate.md) — generate X.509 certificates backed by RSA keys
- [Secret Sharing](secret-sharing.md) — export the key secret to other namespaces
