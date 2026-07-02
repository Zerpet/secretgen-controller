# SSHKey

`SSHKey` generates an SSH key pair and stores it in a Kubernetes `Secret`.

**API:** `secretgen.k14s.io/v1alpha1`  
**Kind:** `SSHKey`

---

## Default Secret

By default the generated `Secret` has:

| Field | Value |
|-------|-------|
| `type` | `kubernetes.io/ssh-auth` |
| `data.ssh-privatekey` | PEM-encoded SSH private key |
| `data.ssh-authorizedkey` | SSH authorized-keys format public key |

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
| `$(privateKey)` | PEM-encoded SSH private key |
| `$(authorizedKey)` | Public key in `authorized_keys` format |

---

## Examples

### Minimal — Default Key Pair

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: SSHKey
metadata:
  name: deploy-key
spec: {}
```

Produces a `Secret` named `deploy-key` of type `kubernetes.io/ssh-auth` with keys `ssh-privatekey` and `ssh-authorizedkey`.

---

### Custom Secret Projection

Store the keys under application-specific names:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: SSHKey
metadata:
  name: git-deploy-key
spec:
  secretTemplate:
    type: Opaque
    stringData:
      id_rsa: $(privateKey)
      id_rsa.pub: $(authorizedKey)
```

---

### Private Key Only (for SSH Client Config)

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: SSHKey
metadata:
  name: ssh-client-key
spec:
  secretTemplate:
    type: Opaque
    stringData:
      ssh-private-key: $(privateKey)
```

---

## Using the Generated Secret

Mount the SSH private key for use by a Git-sync sidecar or CI job:

```yaml
volumes:
- name: ssh-key
  secret:
    secretName: git-deploy-key
    defaultMode: 0400  # restrict permissions — SSH requires private key to not be world-readable
containers:
- name: git-sync
  volumeMounts:
  - name: ssh-key
    mountPath: /root/.ssh
    readOnly: true
```

Use the authorized key to grant access on a remote server by copying `$(authorizedKey)` into `~/.ssh/authorized_keys`.

---

## See Also

- [RSAKey](rsa-key.md) — generate a standalone RSA key pair (not SSH format)
- [Secret Sharing](secret-sharing.md) — export the key secret to other namespaces
