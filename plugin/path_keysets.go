package secretsengine

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-beta-sdk-go"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	graphtrustframework "github.com/microsoftgraph/msgraph-beta-sdk-go/trustframework"
)

func pathKeys(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "keysets/?$",
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ListOperation: &framework.PathOperation{
					Callback: b.pathKeySetsList,
				},
			},
			HelpSynopsis:    "",
			HelpDescription: "",
		},
		{
			Pattern: "keysets/" + framework.GenericNameRegex("id"),
			Fields: map[string]*framework.FieldSchema{
				"id": {
					Type:        framework.TypeString,
					Description: "KeySet ID (eg. B2C_1A_RestApiKey)",
					Required:    true,
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathKeySetsRead,
				},
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathKeySetsCreate,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathKeySetsDelete,
				},
			},
			ExistenceCheck:  b.pathKeySetsExistenceCheck,
			HelpSynopsis:    "",
			HelpDescription: "",
		},
		{
			Pattern: "keysets/" + framework.GenericNameRegex("id") + "/uploadSecret",
			Fields: map[string]*framework.FieldSchema{
				"id": {
					Type:        framework.TypeString,
					Description: "KeySet ID (eg. B2C_1A_RestApiKey)",
					Required:    true,
				},
				"secret": {
					Type:        framework.TypeString,
					Description: "Secret to upload",
					Required:    true,
				},
				"use": {
					Type:        framework.TypeString,
					Description: "Use of the secret",
					Default:     "sig",
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathKeySetsUploadSecret,
				},
			},
			HelpSynopsis:    "",
			HelpDescription: "",
		},
		{
			Pattern: "keysets/" + framework.GenericNameRegex("id") + "/generateKey",
			Fields: map[string]*framework.FieldSchema{
				"id": {
					Type:        framework.TypeString,
					Description: "KeySet ID (eg. B2C_1A_RestApiKey)",
					Required:    true,
				},
				"use": {
					Type:        framework.TypeString,
					Description: "Use of the secret",
					Default:     "sig",
				},
				"kty": {
					Type:        framework.TypeString,
					Description: "Key type",
					Default:     "RSA",
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathKeySetsGenerateKey,
				},
			},
			HelpSynopsis:    "",
			HelpDescription: "",
		},
	}
}

func (b *backend) pathKeySetsRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	client, err := b.GetClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	requestedKey := data.Get("id").(string)
	key, err := b.getKey(ctx, client, requestedKey)
	if err != nil {
		return nil, err
	}

	exp := "n/a"
	if key.Exp > 0 {
		exp = time.Unix(key.Exp, 0).Local().Format("2006-01-02 15:04:05")
	}
	nbf := "n/a"
	if key.Nbf > 0 {
		nbf = time.Unix(key.Nbf, 0).Local().Format("2006-01-02 15:04:05")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"kid": key.Kid,
			"kty": key.Kty,
			"use": key.Use,
			"exp": exp,
			"nbf": nbf,
		},
	}, nil
}

func (b *backend) getKey(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, keyName string) (AzureKey, error) {
	var azureKey AzureKey
	key, err := client.TrustFramework().KeySets().ByTrustFrameworkKeySetId(keyName).GetActiveKey().Get(ctx, nil)
	if err != nil {
		return azureKey, err
	}

	azureKey = AzureKey{
		Kid: *key.GetKid(),
		Kty: *key.GetKty(),
		Use: *key.GetUse(),
	}

	if key.GetNbf() != nil {
		azureKey.Nbf = *key.GetNbf()
	}

	if key.GetExp() != nil {
		azureKey.Exp = *key.GetExp()
	}

	return azureKey, nil
}

func (b *backend) pathKeySetsList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	client, err := b.GetClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	keys, err := b.listKeys(ctx, client)
	if err != nil {
		return nil, err
	}

	var item_list []string

	for _, key := range keys {
		item_list = append(item_list, key.ID)
	}

	return logical.ListResponse(item_list), nil
}

func (b *backend) listKeys(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]AzureKeySet, error) {
	keys, err := client.TrustFramework().KeySets().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var azureKeySets []AzureKeySet
	for _, keySet := range keys.GetValue() {
		var azureKeySet AzureKeySet = AzureKeySet{
			ID: *keySet.GetId(),
		}
		azureKeySets = append(azureKeySets, azureKeySet)
	}

	return azureKeySets, nil
}

func (b *backend) pathKeySetsExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	client, err := b.GetClient(ctx, req.Storage)
	if err != nil {
		return false, err
	}

	requestedKey := data.Get("id").(string)
	key, _ := b.getKey(ctx, client, requestedKey)
	return key.Kid != "", nil
}

func (b *backend) pathKeySetsCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	id := data.Get("id").(string)
	// If ID not provided or not in the for of B2C_1A_.*, return error
	if id == "" {
		return logical.ErrorResponse("ID is required"), nil
	} else if !strings.HasPrefix(id, "B2C_1A_") {
		return logical.ErrorResponse("ID must start with B2C_1A_"), nil
	}

	client, err := b.GetClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	requestBody := graphmodels.NewTrustFrameworkKeySet()
	requestBody.SetId(&id)

	_, err = client.TrustFramework().KeySets().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathKeySetsUploadSecret(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	id := data.Get("id").(string)
	k := data.Get("secret").(string)
	use := data.Get("use").(string)

	// If ID not provided or not in the for of B2C_1A_.*, return error
	if !strings.HasPrefix(id, "B2C_1A_") {
		return logical.ErrorResponse("ID must start with B2C_1A_"), nil
	}

	client, err := b.GetClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	requestBody := graphtrustframework.NewKeySetsItemUploadSecretPostRequestBody()
	requestBody.SetUse(&use)
	requestBody.SetK(&k)
	nbf := time.Now().Unix()
	requestBody.SetNbf(&nbf)
	exp := time.Now().Add(time.Hour * 24 * 365).Unix()
	requestBody.SetExp(&exp)

	_, err = client.TrustFramework().KeySets().ByTrustFrameworkKeySetId(id).UploadSecret().Post(context.Background(), requestBody, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathKeySetsDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	id := data.Get("id").(string)

	// If ID not provided or not in the for of B2C_1A_.*, return error
	if !strings.HasPrefix(id, "B2C_1A_") {
		return logical.ErrorResponse("ID must start with B2C_1A_"), nil
	}

	client, err := b.GetClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	err = client.TrustFramework().KeySets().ByTrustFrameworkKeySetId(id).Delete(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathKeySetsGenerateKey(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	id := data.Get("id").(string)
	use := data.Get("use").(string)
	kty := data.Get("kty").(string)

	// If ID not provided or not in the for of B2C_1A_.*, return error
	if !strings.HasPrefix(id, "B2C_1A_") {
		return logical.ErrorResponse("ID must start with B2C_1A_"), nil
	}

	client, err := b.GetClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	requestBody := graphtrustframework.NewKeySetsItemGenerateKeyPostRequestBody()
	requestBody.SetUse(&use)
	requestBody.SetKty(&kty)
	nbf := time.Now().Unix()
	requestBody.SetNbf(&nbf)
	exp := time.Now().Add(time.Hour * 24 * 365).Unix()
	requestBody.SetExp(&exp)

	_, err = client.TrustFramework().KeySets().ByTrustFrameworkKeySetId(id).GenerateKey().Post(context.Background(), requestBody, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type AzureKeySet struct {
	ID   string     `json:"id"`
	Keys []AzureKey `json:"keys"`
}

type AzureKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	Exp int64  `json:"exp"`
	Nbf int64  `json:"nbf"`
}
