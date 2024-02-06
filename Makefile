.PHONY: all

all: login init test

login:
	vault login root

init:
	vault secrets enable --path="b2c" aadb2c
	vault write b2c/config @b2c-config.json

test:
	vault read b2c/config
	vault list b2c/keysets
	vault write -force b2c/keysets/B2C_1A_ApiKey1
	vault write -force b2c/keysets/B2C_1A_ApiKey2
	vault write b2c/keysets/B2C_1A_ApiKey1/uploadSecret secret=HelloWorld123
	vault write -force b2c/keysets/B2C_1A_ApiKey2/generateKey
	vault list b2c/keysets
	vault read b2c/keysets/B2C_1A_ApiKey1
	vault read b2c/keysets/B2C_1A_ApiKey2
	vault delete b2c/keysets/B2C_1A_ApiKey1
	vault delete b2c/keysets/B2C_1A_ApiKey2
	vault list b2c/keysets

clean:
	vault delete b2c/keysets/B2C_1A_ApiKey1
	vault delete b2c/keysets/B2C_1A_ApiKey2
	vault secrets disable b2c