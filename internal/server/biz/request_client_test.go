package biz

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent/project"
	"github.com/looplj/axonhub/internal/ent/request"
	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/httpclient"
)

func TestRequestService_CreateRequestPersistsClient(t *testing.T) {
	svc, client, ctx := setupTestRequestService(t)
	defer client.Close()

	proj, err := client.Project.Create().
		SetName("request-client-project").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	ctx = contexts.WithProjectID(ctx, proj.ID)

	tests := []struct {
		name   string
		client llm.RequestClient
		want   request.Client
	}{
		{
			name:   "codex",
			client: llm.RequestClientCodex,
			want:   request.ClientCodex,
		},
		{
			name: "empty defaults to unknown",
			want: request.ClientUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := svc.CreateRequest(
				ctx,
				&llm.Request{Model: "glm-5.2", Client: tt.client},
				&httpclient.Request{JSONBody: []byte(`{"model":"glm-5.2"}`)},
				llm.APIFormatOpenAIResponse,
			)
			require.NoError(t, err)
			require.Equal(t, tt.want, created.Client)
		})
	}
}
