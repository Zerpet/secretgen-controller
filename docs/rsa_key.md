### RSA Key

Please let us know in issues what kind of configurability is wanted.

`spec` fields:

- [`secretTemplate`](secret-template-field.md)

#### Secret Template

Available variables:

- `$(publicKey)`
- `$(privateKey)`

#### Example

```
apiVersion: secretgen.k14s.io/v1alpha1
kind: RSAKey
metadata:
  name: rsa-key
spec: {}
```

```bash
sgctl rsa-key create my-rsa-key
sgctl rsa-key describe my-rsa-key
kubectl get secret my-rsa-key
```
