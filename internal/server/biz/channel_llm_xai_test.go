package biz

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/ent/enttest"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/transformer/xai"
)

func TestXAIChannel_UsesImportedOAuthCredentials(t *testing.T) {
	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=0")
	defer client.Close()

	ctx := authz.WithTestBypass(context.Background())
	entChannel := client.Channel.Create().
		SetName("Grok OAuth Channel").
		SetType(channel.TypeXai).
		SetBaseURL(xai.OAuthBaseURL).
		SetCredentials(objects.ChannelCredentials{
			APIKey: `{"access_token":"grok-access","refresh_token":"grok-refresh","client_id":"grok-cli-client","expires_at":"2099-01-01T00:00:00Z"}`,
		}).
		SetSupportedModels([]string{"grok-4"}).
		SetDefaultTestModel("grok-4").
		SaveX(ctx)

	channelSvc := NewChannelServiceForTest(client)
	built, err := channelSvc.buildChannelWithTransformer(entChannel)
	require.NoError(t, err)

	req, err := built.Outbound.TransformRequest(context.Background(), &llm.Request{
		Model: "grok-4",
		Messages: []llm.Message{
			{Role: "user", Content: llm.MessageContent{Content: lo.ToPtr("hello")}},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, req.Auth)
	require.Equal(t, httpclient.AuthTypeBearer, req.Auth.Type)
	require.Equal(t, "grok-access", req.Auth.APIKey)
}
