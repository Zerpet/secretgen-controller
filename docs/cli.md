# sgctl CLI Reference

`sgctl` is the command-line interface for managing secretgen-controller resources in Kubernetes. It provides commands to create, update, delete, and query Certificates, Passwords, RSA Keys, SSH Keys, SecretExports, SecretImports, and SecretTemplates.

## Global Flags

The following flags are available on all `sgctl` commands:

- `--kubeconfig` (string) - Path to kubeconfig file (defaults to KUBECONFIG env or ~/.kube/config)
- `--context` (string) - Kubernetes context to use
- `-n, --namespace` (string) - Kubernetes namespace (defaults to current context namespace)
- `-o, --output` (string) - Output format: `table`, `json`, `yaml` (default: `table`)

## Commands

### version

Print sgctl version

```bash
sgctl version
```

### certificate

Manage Certificate resources. Aliases: `cert`, `certificates`

#### Subcommands

##### create

Create a Certificate resource.

**Usage:**
```bash
sgctl certificate create <name> [flags]
```

**Flags:**
- `--is-ca` (boolean) - Whether this is a CA certificate (default: false)
- `--ca-ref` (string) - Name of CA Certificate to sign with
- `--common-name` (string) - Certificate Common Name (CN)
- `--organization` (string) - Certificate Organization field
- `--alt-names` (strings) - Comma-separated Subject Alternative Names (IPs or DNS names)
- `--ext-key-usage` (strings) - Comma-separated Extended Key Usage values (client_auth, server_auth)
- `--duration` (int) - Certificate validity in days (0 = controller default of 365)
- `--secret-template-file` (string) - Path to YAML file defining a secretTemplate override

**Examples:**

Create a self-signed CA certificate:
```bash
sgctl certificate create my-ca --is-ca --common-name "My CA"
```

Create a certificate signed by a CA:
```bash
sgctl certificate create my-cert --ca-ref my-ca --common-name "example.com" --alt-names "example.com,*.example.com"
```

Create a certificate with extended key usage:
```bash
sgctl certificate create server-cert --ca-ref my-ca --common-name "server.example.com" --ext-key-usage "server_auth"
```

Create a certificate with custom validity period:
```bash
sgctl certificate create short-lived --duration 30 --common-name "temporary.example.com"
```

##### update

Update a Certificate resource.

**Usage:**
```bash
sgctl certificate update <name> [flags]
```

**Flags:** Same as `create`

**Examples:**

Update certificate to add alternative names:
```bash
sgctl certificate update my-cert --alt-names "example.com,www.example.com,api.example.com"
```

Update certificate validity period:
```bash
sgctl certificate update my-cert --duration 730
```

##### delete

Delete a Certificate resource.

**Usage:**
```bash
sgctl certificate delete <name> [flags]
```

**Flags:**
- `-y, --yes` (boolean) - Skip confirmation prompt (default: false)

**Examples:**

Delete a certificate with confirmation:
```bash
sgctl certificate delete my-cert
```

Delete a certificate without confirmation:
```bash
sgctl certificate delete my-cert --yes
```

##### get

Get a Certificate resource.

**Usage:**
```bash
sgctl certificate get <name> [flags]
```

**Examples:**

Get certificate in table format (default):
```bash
sgctl certificate get my-cert
```

Get certificate in JSON format:
```bash
sgctl certificate get my-cert --output json
```

Get certificate in YAML format:
```bash
sgctl certificate get my-cert --output yaml
```

##### describe

Describe a Certificate resource with detailed information.

**Usage:**
```bash
sgctl certificate describe <name> [flags]
```

**Examples:**

```bash
sgctl certificate describe my-cert
```

##### list

List all Certificate resources in the current namespace.

**Usage:**
```bash
sgctl certificate list [flags]
```

**Examples:**

List all certificates in default table format:
```bash
sgctl certificate list
```

List all certificates in JSON format:
```bash
sgctl certificate list --output json
```

List all certificates in a specific namespace:
```bash
sgctl certificate list --namespace kube-system
```

---

### password

Manage Password resources. Aliases: `pwd`, `passwords`

#### Subcommands

##### create

Create a Password resource.

**Usage:**
```bash
sgctl password create <name> [flags]
```

**Flags:**
- `--length` (int) - Total password length (controller default: 40)
- `--digits` (int) - Number of digit characters
- `--symbols` (int) - Number of symbol characters
- `--uppercase` (int) - Number of uppercase letters
- `--lowercase` (int) - Number of lowercase letters
- `--symbol-charset` (string) - Set of symbols to use (controller default: !@#$%&*;.:)
- `--secret-template-file` (string) - Path to YAML file defining a secretTemplate override

**Examples:**

Create a password with default settings:
```bash
sgctl password create my-password
```

Create a password with custom length:
```bash
sgctl password create long-password --length 64
```

Create a password with specific character requirements:
```bash
sgctl password create complex-password --length 32 --digits 5 --symbols 3 --uppercase 8 --lowercase 8
```

Create a password with custom symbol set:
```bash
sgctl password create special-password --length 20 --symbols 5 --symbol-charset "!@#$%"
```

##### update

Update a Password resource.

**Usage:**
```bash
sgctl password update <name> [flags]
```

**Flags:** Same as `create`

**Examples:**

Update password length:
```bash
sgctl password update my-password --length 50
```

Update password character requirements:
```bash
sgctl password update my-password --digits 8 --symbols 4
```

##### delete

Delete a Password resource.

**Usage:**
```bash
sgctl password delete <name> [flags]
```

**Flags:**
- `-y, --yes` (boolean) - Skip confirmation prompt (default: false)

**Examples:**

Delete a password with confirmation:
```bash
sgctl password delete my-password
```

Delete a password without confirmation:
```bash
sgctl password delete my-password --yes
```

##### get

Get a Password resource.

**Usage:**
```bash
sgctl password get <name> [flags]
```

**Examples:**

```bash
sgctl password get my-password
sgctl password get my-password --output json
```

##### describe

Describe a Password resource with detailed information.

**Usage:**
```bash
sgctl password describe <name> [flags]
```

**Examples:**

```bash
sgctl password describe my-password
```

##### list

List all Password resources in the current namespace.

**Usage:**
```bash
sgctl password list [flags]
```

**Examples:**

```bash
sgctl password list
sgctl password list --namespace kube-system
```

---

### rsa-key

Manage RSAKey resources. Aliases: `rsakey`, `rsa-keys`

#### Subcommands

##### create

Create an RSAKey resource.

**Usage:**
```bash
sgctl rsa-key create <name> [flags]
```

**Flags:**
- `--secret-template-file` (string) - Path to YAML file defining a secretTemplate override

**Examples:**

Create an RSA key:
```bash
sgctl rsa-key create my-rsa-key
```

##### update

Update an RSAKey resource.

**Usage:**
```bash
sgctl rsa-key update <name> [flags]
```

**Flags:**
- `--secret-template-file` (string) - Path to YAML file defining a secretTemplate override

**Examples:**

```bash
sgctl rsa-key update my-rsa-key
```

##### delete

Delete an RSAKey resource.

**Usage:**
```bash
sgctl rsa-key delete <name> [flags]
```

**Flags:**
- `-y, --yes` (boolean) - Skip confirmation prompt (default: false)

**Examples:**

```bash
sgctl rsa-key delete my-rsa-key --yes
```

##### get

Get an RSAKey resource.

**Usage:**
```bash
sgctl rsa-key get <name> [flags]
```

**Examples:**

```bash
sgctl rsa-key get my-rsa-key
```

##### describe

Describe an RSAKey resource with detailed information.

**Usage:**
```bash
sgctl rsa-key describe <name> [flags]
```

**Examples:**

```bash
sgctl rsa-key describe my-rsa-key
```

##### list

List all RSAKey resources in the current namespace.

**Usage:**
```bash
sgctl rsa-key list [flags]
```

**Examples:**

```bash
sgctl rsa-key list
```

---

### ssh-key

Manage SSHKey resources. Aliases: `sshkey`, `ssh-keys`

#### Subcommands

##### create

Create an SSHKey resource.

**Usage:**
```bash
sgctl ssh-key create <name> [flags]
```

**Flags:**
- `--secret-template-file` (string) - Path to YAML file defining a secretTemplate override

**Examples:**

Create an SSH key:
```bash
sgctl ssh-key create my-ssh-key
```

##### update

Update an SSHKey resource.

**Usage:**
```bash
sgctl ssh-key update <name> [flags]
```

**Flags:**
- `--secret-template-file` (string) - Path to YAML file defining a secretTemplate override

**Examples:**

```bash
sgctl ssh-key update my-ssh-key
```

##### delete

Delete an SSHKey resource.

**Usage:**
```bash
sgctl ssh-key delete <name> [flags]
```

**Flags:**
- `-y, --yes` (boolean) - Skip confirmation prompt (default: false)

**Examples:**

```bash
sgctl ssh-key delete my-ssh-key --yes
```

##### get

Get an SSHKey resource.

**Usage:**
```bash
sgctl ssh-key get <name> [flags]
```

**Examples:**

```bash
sgctl ssh-key get my-ssh-key
```

##### describe

Describe an SSHKey resource with detailed information.

**Usage:**
```bash
sgctl ssh-key describe <name> [flags]
```

**Examples:**

```bash
sgctl ssh-key describe my-ssh-key
```

##### list

List all SSHKey resources in the current namespace.

**Usage:**
```bash
sgctl ssh-key list [flags]
```

**Examples:**

```bash
sgctl ssh-key list
```

---

### secret-export

Manage SecretExport resources. Aliases: `secexp`, `secret-exports`

#### Subcommands

##### create

Create a SecretExport resource.

**Usage:**
```bash
sgctl secret-export create <name> [flags]
```

**Flags:**
- `--to-namespace` (string) - Single destination namespace (use * for all)
- `--to-namespaces` (strings) - Comma-separated list of destination namespaces
- `--to-selector` (string) - JSON array of namespace selector match fields, e.g. `'[{"key":"metadata.labels.env","operator":"In","values":["prod"]}]'`

**Examples:**

Export secret to a single namespace:
```bash
sgctl secret-export create my-export --to-namespace target-namespace
```

Export secret to all namespaces:
```bash
sgctl secret-export create my-export --to-namespace "*"
```

Export secret to multiple specific namespaces:
```bash
sgctl secret-export create my-export --to-namespaces namespace1,namespace2,namespace3
```

Export secret to namespaces matching a selector:
```bash
sgctl secret-export create my-export --to-selector '[{"key":"metadata.labels.env","operator":"In","values":["prod"]}]'
```

##### update

Update a SecretExport resource.

**Usage:**
```bash
sgctl secret-export update <name> [flags]
```

**Flags:** Same as `create`

**Examples:**

```bash
sgctl secret-export update my-export --to-namespaces namespace1,namespace2
```

##### delete

Delete a SecretExport resource.

**Usage:**
```bash
sgctl secret-export delete <name> [flags]
```

**Flags:**
- `-y, --yes` (boolean) - Skip confirmation prompt (default: false)

**Examples:**

```bash
sgctl secret-export delete my-export --yes
```

##### get

Get a SecretExport resource.

**Usage:**
```bash
sgctl secret-export get <name> [flags]
```

**Examples:**

```bash
sgctl secret-export get my-export
```

##### describe

Describe a SecretExport resource with detailed information.

**Usage:**
```bash
sgctl secret-export describe <name> [flags]
```

**Examples:**

```bash
sgctl secret-export describe my-export
```

##### list

List all SecretExport resources in the current namespace.

**Usage:**
```bash
sgctl secret-export list [flags]
```

**Examples:**

```bash
sgctl secret-export list
```

---

### secret-import

Manage SecretImport resources. Aliases: `secimp`, `secret-imports`

#### Subcommands

##### create

Create a SecretImport resource.

**Usage:**
```bash
sgctl secret-import create <name> [flags]
```

**Flags:**
- `--from-namespace` (string) - Source namespace to import secret from

**Examples:**

Import a secret from another namespace:
```bash
sgctl secret-import create my-import --from-namespace source-namespace
```

##### update

Update a SecretImport resource.

**Usage:**
```bash
sgctl secret-import update <name> [flags]
```

**Flags:** Same as `create`

**Examples:**

```bash
sgctl secret-import update my-import --from-namespace different-namespace
```

##### delete

Delete a SecretImport resource.

**Usage:**
```bash
sgctl secret-import delete <name> [flags]
```

**Flags:**
- `-y, --yes` (boolean) - Skip confirmation prompt (default: false)

**Examples:**

```bash
sgctl secret-import delete my-import --yes
```

##### get

Get a SecretImport resource.

**Usage:**
```bash
sgctl secret-import get <name> [flags]
```

**Examples:**

```bash
sgctl secret-import get my-import
```

##### describe

Describe a SecretImport resource with detailed information.

**Usage:**
```bash
sgctl secret-import describe <name> [flags]
```

**Examples:**

```bash
sgctl secret-import describe my-import
```

##### list

List all SecretImport resources in the current namespace.

**Usage:**
```bash
sgctl secret-import list [flags]
```

**Examples:**

```bash
sgctl secret-import list
```

---

### secret-template

Manage SecretTemplate resources. Aliases: `tmpl`, `secret-templates`

#### Subcommands

##### create

Create a SecretTemplate resource.

**Usage:**
```bash
sgctl secret-template create <name> [flags]
```

**Flags:**
- `--from-file` (string) - Path to YAML file defining the secretTemplate

**Examples:**

Create a SecretTemplate from a file:
```bash
sgctl secret-template create my-template --from-file template.yaml
```

##### update

Update a SecretTemplate resource.

**Usage:**
```bash
sgctl secret-template update <name> [flags]
```

**Flags:** Same as `create`

**Examples:**

```bash
sgctl secret-template update my-template --from-file template.yaml
```

##### delete

Delete a SecretTemplate resource.

**Usage:**
```bash
sgctl secret-template delete <name> [flags]
```

**Flags:**
- `-y, --yes` (boolean) - Skip confirmation prompt (default: false)

**Examples:**

```bash
sgctl secret-template delete my-template --yes
```

##### get

Get a SecretTemplate resource.

**Usage:**
```bash
sgctl secret-template get <name> [flags]
```

**Examples:**

```bash
sgctl secret-template get my-template
```

##### describe

Describe a SecretTemplate resource with detailed information.

**Usage:**
```bash
sgctl secret-template describe <name> [flags]
```

**Examples:**

```bash
sgctl secret-template describe my-template
```

##### list

List all SecretTemplate resources in the current namespace.

**Usage:**
```bash
sgctl secret-template list [flags]
```

**Examples:**

```bash
sgctl secret-template list
```

---

## Output Formats

All commands support three output formats via the `-o` or `--output` flag:

### table

Default format with human-readable table output:

```bash
sgctl password list
```

Output:
```
NAME             LENGTH  DESCRIPTION                      AGE
my-password      40      Password generated successfully  5d
long-password    64      Password generated successfully  2d
```

### json

JSON format for programmatic processing:

```bash
sgctl password list --output json
```

### yaml

YAML format for resource definitions:

```bash
sgctl password list --output yaml
```

## Common Examples

### Working with Multiple Namespaces

Get a resource from a specific namespace:

```bash
sgctl certificate get my-cert --namespace kube-system
```

List resources from a specific namespace:

```bash
sgctl password list --namespace production
```

### Using Kubeconfig

Specify a custom kubeconfig file:

```bash
sgctl certificate list --kubeconfig /path/to/kubeconfig
```

Use a specific context:

```bash
sgctl password list --context my-cluster-context
```

### Combining Flags

```bash
sgctl certificate get my-cert --namespace kube-system --context my-context --output json
```

### Retrieving Secret Values

Get generated password value:

```bash
kubectl get secret my-password -o jsonpath='{.data.password}' | base64 -d
```

Get certificate from generated secret:

```bash
kubectl get secret my-cert -o jsonpath='{.data.tls\.crt}' | base64 -d
```

Get SSH public key:

```bash
kubectl get secret my-ssh-key -o jsonpath='{.data.ssh-publickey}' | base64 -d
```
