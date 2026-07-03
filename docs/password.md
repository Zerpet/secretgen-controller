### Password

Password CRD allows to generate alpha-numeric passwords. The default charset includes lowercase characters, uppercase characters and digits.

`spec` fields:

- `length` (optional; default is 40) number of characters
- [`secretTemplate`](secret-template-field.md)
- `digits` (optional; default is 0) minimun number of digits in the generated password
- `lowercaseLetters` (optional; default is 0) minimun number of lowercase letters in the generated password
- `uppercaseLetters` (optional; default is 0) minimun number of uppercase letters in the generated password 
- `symbols` (optional; default is 0) minimun number of symbols in the generated password 
- `symbolCharSet`(optional, string) list of characters available when generating a password with symbols 


#### Secret Template

Available variables:

- `$(value)`

#### Examples

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: long-user-password
spec:
  length: 124
```

```bash
sgctl password create long-user-password --length 124
sgctl password describe long-user-password
kubectl get secret long-user-password -o jsonpath='{.data.password}' | base64 -d
```

With custom secret projection:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: long-user-password
spec:
  length: 124
  secretTemplate:
    type: Opaque
    stringData:
      postgresql-password: $(value)
```

```bash
kubectl apply -f password.yml
sgctl password describe long-user-password
kubectl get secret long-user-password -o jsonpath='{.data.postgresql-password}' | base64 -d
```

With custom specification:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: custom-user-password
spec:
  length: 27    
  digits: 2
  uppercaseLetters: 4
  lowercaseLetters: 10
  symbols: 3  
```

```bash
sgctl password create custom-user-password --length 27 --digits 2 --uppercase 4 --lowercase 10 --symbols 3
sgctl password describe custom-user-password
```

With only symbols and specific symbol charset specification:

```yaml
apiVersion: secretgen.k14s.io/v1alpha1
kind: Password
metadata:
  name: only-symbols-user-password
spec:      
  length: 7
  symbols: 7  
  symbolCharSet: "!$#%"
```

```bash
sgctl password create only-symbols-user-password --length 7 --symbols 7 --symbol-charset "!$#%"
kubectl get secret only-symbols-user-password -o jsonpath='{.data.password}' | base64 -d
```