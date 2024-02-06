# Vault Plugin: Azure AD B2C Secrets Backend

This is a standalone backend plugin for use with [Hashicorp Vault](https://www.github.com/hashicorp/vault).
This plugin manages Policy Keys for [Microsoft Azure AD B2C](https://learn.microsoft.com/en-us/azure/active-directory-b2c/overview).

## Getting Started

This is a [Vault plugin](https://developer.hashicorp.com/vault/docs/plugins)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to [get started with Vault](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-install).

To learn specifically about how plugins work, see documentation on [Vault plugins](https://developer.hashicorp.com/vault/docs/plugins).

You can start HashiCorp Vault server in [development mode](https://developer.hashicorp.com/vault/docs/concepts/dev-server) to demonstrate and evaluate the secrets engine. Vault starts unsealed in this configuration, and you do not need to register the plugin.

```sh
vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins -log-level=debug
export VAULT_ADDR='http://localhost:8200'
```

> **Warning:** Running Vault in development mode is useful for evaluating the plugin, but should **never** be used in production.

## Usage

### Build the binary

```sh
# Clone this repository
git clone https://github.com/justinwilloughby/vault-plugin-secrets-azureadb2c.git
# Navigate to the directory
cd vault-plugin-secrets-azureadb2c
# If running vault in dev mode, I find it useful to make a vault/plugins directory here to store my plugin
mkdir -p vault/plugins
# Build the binary
go build -o ./vault/plugins/aadb2c ./main.go
```

### Enable and configure the plugin

Enable the `aadb2c` secrets engine at the `b2c/` path:

```sh
vault secrets enable --path="b2c" aadb2c
```

Write the configuration data to `b2c/config` in a single command (assuming the `AZURE_` environment variables have been set):

```sh
vault write b2c/config \
  client_id=$AZURE_CLIENT_ID \
  client_secret=$AZURE_CLIENT_SECRET \
  tenant_id=$AZURE_TENANT_ID \
```

Alternatively, create a JSON file. For example, save the following as `b2c-config.json`:

```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "tenant_id": "your_tenant_id"
}
```

Write the data to the `b2c/config` path using this file to configure the secrets engine.

```sh
vault write b2c/config @b2c-config.json
```

### Commands

#### List keysets

Returns the names of the keysets in the tenant:

```sh
vault list b2c/keysets
```

#### Read keyset active key

Returns the active key for the keyset:

```sh
vault read b2c/keysets/B2C_1A_ApiKey
```

#### Create keyset

```sh
vault write -force b2c/keysets/B2C_1A_ApiKey
```

#### Upload secret to keyset

```sh
vault write b2c/keysets/B2C_1A_ApiKey/uploadSecret secret=HelloWorld123!
```

#### Generate key for keyset

```sh
vault write b2c/keysets/B2C_1A_ApiKey/generateKey
```

#### Delete keyset


```sh
vault delete b2c/keysets/B2C_1A_ApiKey
```

### Using Makefile to configure and test

After the vault server is running, there is a Makefile that automates the commands shown above to runthrough a quick setup and functional test.

You can run it as soon as you start the Vault server in dev mode and set the VAULT_ADDR. Make sure you're in the root of the directory.

```sh
make
```

This will do a vault login, configure the secrets engine a b2c-config.json file you create, and then run through each of the commands.