# Quickstart

## Prerequisites

- A basic understanding of [HashiCorp Vault](https://www.hashicorp.com/products/vault) (see [What is Vault?](https://developer.hashicorp.com/vault/docs/what-is-vault) for details).
- A [HashiCorp Vault server](https://developer.hashicorp.com/vault/docs/install).
  - **Note:** This guide also includes [quick start instructions](#quick-start-for-evaluation) for running a Vault server in development mode if you'd like to evaluate this plugin before starting to use it.
- The [HashiCorp Vault CLI](https://developer.hashicorp.com/vault/downloads) installed on your device.
- [Go](https://go.dev/doc/install) (if you want to build the plugin from source).

## Quick start (for evaluation)

You can start HashiCorp Vault server in [development mode](https://developer.hashicorp.com/vault/docs/concepts/dev-server) to demonstrate and evaluate the secrets engine. Vault starts unsealed in this configuration, and you do not need to register the plugin.

```sh
vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins -log-level=debug
```

> **Warning:** Running Vault in development mode is useful for evaluating the plugin, but should **never** be used in production.

Connect to the Vault server in a **new** terminal to [enable the secrets engine](#enable-and-configure-the-plugin) and start using it.

## Getting started

### Build the binary

```sh
# Clone this repository
git clone https://github.com/justinwilloughby/vault-plugin-secrets-azureadb2c.git
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
  tenant_id=AZURE_$TENANT_ID \
  subscription_id=$AZURE_SUBSCRIPTION_ID
```

Alternatively, create a JSON file. For example, save the following as `b2c-config.json`:

```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "tenant_id": "your_tenant_id",
  "subscription_id": "your_subscription_id"
}
```

Write the data to the `b2c/config` path using this file to configure the secrets engine.

```sh
vault write b2c/config @b2c-config.json
```

## Usage

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

#### Delete item

Delete the specified item:

```sh
vault delete b2c/keysets/B2C_1A_ApiKey
```