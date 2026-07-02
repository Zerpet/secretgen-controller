# Password

`Password` generates a random alpha-numeric password and stores it in a Kubernetes `Secret`.

**API:** `secretgen.k14s.io/v1alpha1`  
**Kind:** `Password`

---

## Default Secret

By default the generated `Secret` has:

| Field | Value |
|-------|-------|
| `type` | `kubernetes.io/basic-auth` |
| `data.password` | base64-encoded password string |

---

## Spec Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `length` | `int` | `40` | Total number of characters in the password |
| `digits` | `int` | `0` | Minimum number of digit characters (`0`–`9`) |
| `lowercaseLetters` | `int` | `0` | Minimum number of lowercase letters |
| `uppercaseLetters` | `int` | `0` | Minimum number of uppercase letters |
| `symbols` | `int` | `0` | Minimum number of symbol characters |
| `symbolCharSet` | `string` | `!@#$%&*;.:` | Characters to draw symbols from |
| `secretTemplate` | object | — | Customise the generated `Secret` (see below) |

> The sum of `digits`, `lowercaseLetters`, `uppercaseLetters`, and `symbols` must not exceed `length`. Remaining characters are drawn from the full default charset (lowercase + uppercase + digits).

### `secretTemplate` Sub-fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | `string` | Kubernetes `Secret` type |
| `stringData` | `map[string]string` | Key/value pairs; values may reference `$(value)` |

**Available variables:**

| Variable | Description |
|----------|-------------|
| `$(value)` | The generated password string |

---

## Examples

### Minimal — Default Length (40 Characters)

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: user-password
spec: {}
```

Produces a `Secret` of type `kubernetes.io/basic-auth` with a 40-character password in `data.password`.

---

### Custom Length

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: long-api-key
spec:
  length: 128
```

---

### Custom Secret Projection (Opaque, Custom Key)

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

Produces:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pg-password
type: Opaque
data:
  postgresql-password: <base64 encoded password>
```

---

### Controlled Character Mix

Generate a 27-character password that contains at least 2 digits, 4 uppercase letters, 10 lowercase letters, and 3 symbols:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: complex-password
spec:
  length: 27
  digits: 2
  uppercaseLetters: 4
  lowercaseLetters: 10
  symbols: 3
```

---

### Symbol-Only Password with Custom Symbol Charset

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: pin-code
spec:
  length: 7
  symbols: 7
  symbolCharSet: "!$#%"
```

---

### Multiple Passwords in One Apply

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: app-session-secret
spec:
  length: 64
---
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: app-db-password
spec:
  length: 32
  secretTemplate:
    type: Opaque
    stringData:
      DB_PASSWORD: $(value)
```

---

## Using the Generated Secret

Reference the secret in a `Pod` or `Deployment` as a normal Kubernetes secret:

```yaml
env:
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: pg-password
      key: postgresql-password
```

Or as a mounted volume:

```yaml
volumes:
- name: pg-pass
  secret:
    secretName: pg-password
```

---

## See Also

- [Concepts — Secret Templates](concepts.md#secret-templates-secrettemplate-field)
- [Secret Sharing](secret-sharing.md) — export this secret to other namespaces
