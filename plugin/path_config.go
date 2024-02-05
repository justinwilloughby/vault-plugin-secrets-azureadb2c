package secretsengine

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type AzureConfig struct {
	SubscriptionID string `json:"subscription_id"`
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
}

func pathConfig(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"subscription_id": {
				Type:        framework.TypeString,
				Description: "Azure Subscription ID",
			},
			"tenant_id": {
				Type:        framework.TypeString,
				Description: "Azure Tenant ID",
			},
			"client_id": {
				Type:        framework.TypeString,
				Description: "Azure Client ID",
			},
			"client_secret": {
				Type:        framework.TypeString,
				Description: "Azure Client Secret",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
		},
		HelpSynopsis:    pathConfigHelpSyn,
		HelpDescription: pathConfigHelpDesc,
	}
}

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = new(AzureConfig)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"subscription_id": config.SubscriptionID,
			"tenant_id":       config.TenantID,
			"client_id":       config.ClientID,
			"client_secret":   config.ClientSecret,
		},
	}, nil
}

func (b *backend) getConfig(ctx context.Context, s logical.Storage) (*AzureConfig, error) {
	entry, err := s.Get(ctx, "config")
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	config := new(AzureConfig)
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = new(AzureConfig)
	}

	if subscriptionID, ok := data.GetOk("subscription_id"); ok {
		config.SubscriptionID = subscriptionID.(string)
	}

	if tenantID, ok := data.GetOk("tenant_id"); ok {
		config.TenantID = tenantID.(string)
	}

	if clientID, ok := data.GetOk("client_id"); ok {
		config.ClientID = clientID.(string)
	}

	if clientSecret, ok := data.GetOk("client_secret"); ok {
		config.ClientSecret = clientSecret.(string)
	}

	entry, err := logical.StorageEntryJSON("config", config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

const pathConfigHelpSyn = `
Configure the Azure AD B2C secrets engine.
`

const pathConfigHelpDesc = `
This endpoint configures the Azure AD B2C secrets engine.

The Azure AD B2C secrets engine requires a valid Azure AD B2C application
registration. The application registration must have the following permissions:

- TrustFrameworkKeyset.ReadWrite.All
`
