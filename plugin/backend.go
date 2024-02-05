package secretsengine

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-beta-sdk-go"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {

	b := Backend()

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

func Backend() *backend {
	var b = &backend{}

	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		Help:        "",
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
			},
		},
		Paths: framework.PathAppend(
			pathKeys(b),
			[]*framework.Path{
				pathConfig(b),
			},
		),
		Secrets: []*framework.Secret{},
	}

	return b
}

type backend struct {
	*framework.Backend

	client *msgraphsdkgo.GraphServiceClient
}

func (b *backend) GetClient(ctx context.Context, s logical.Storage) (*msgraphsdkgo.GraphServiceClient, error) {
	if b.client != nil {
		return b.client, nil
	}

	config, err := b.getConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewClientSecretCredential(config.TenantID, config.ClientID, config.ClientSecret, nil)
	if err != nil {
		return nil, err
	}

	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, err
	}

	b.client = client

	return client, nil
}
