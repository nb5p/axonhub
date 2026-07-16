package codex

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/oauth"
	"github.com/looplj/axonhub/llm/transformer/openai/responses"
)

func TestAlphaSearchInboundTransformer_RoundTrip(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"model":"gpt-5.6-sol","commands":{"search_query":[{"q":"test"}]}}`)
	rawRequest := &httpclient.Request{
		Method: http.MethodPost,
		Path:   "/v1/alpha/search",
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
			SessionHeader:  []string{"session-123"},
		},
		Body: body,
	}

	inbound := NewAlphaSearchInboundTransformer()
	request, err := inbound.TransformRequest(ctx, rawRequest)
	require.NoError(t, err)
	require.Equal(t, "gpt-5.6-sol", request.Model)
	require.Equal(t, llm.RequestTypeAlphaSearch, request.RequestType)
	require.Equal(t, llm.APIFormatOpenAICodexAlphaSearch, request.APIFormat)

	outbound, err := NewOutboundTransformer(Params{
		BaseURL: "wss://chatgpt.com/backend-api/codex#",
		TokenProvider: staticTokenGetter{creds: &oauth.OAuthCredentials{
			AccessToken: testAccessTokenWithAccountID(t),
			ExpiresAt:   time.Now().Add(time.Hour),
		}},
	})
	require.NoError(t, err)

	request.RawRequest = rawRequest
	providerRequest, err := outbound.TransformRequest(ctx, request)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, providerRequest.Method)
	require.Equal(t, "https://chatgpt.com/backend-api/codex/alpha/search", providerRequest.URL)
	require.True(t, bytes.Equal(body, providerRequest.Body))
	require.Equal(t, "application/json", providerRequest.Headers.Get("Accept"))
	require.Equal(t, "session-123", providerRequest.Headers.Get(SessionHeaderHyphen))
	require.Equal(t, llm.RequestTypeAlphaSearch.String(), providerRequest.RequestType)
	require.Equal(t, llm.APIFormatOpenAICodexAlphaSearch.String(), providerRequest.APIFormat)

	providerRequest = httpclient.MergeInboundRequest(providerRequest, rawRequest)
	require.Empty(t, providerRequest.Headers.Get(SessionHeader))

	providerResponse := &httpclient.Response{
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"application/json"}, "X-Test": []string{"preserved"}},
		Body:       []byte(`{"type":"computer_initialize_state","id":"search-1"}`),
		Request:    providerRequest,
	}
	response, err := outbound.TransformResponse(ctx, providerResponse)
	require.NoError(t, err)

	clientResponse, err := inbound.TransformResponse(ctx, response)
	require.NoError(t, err)
	require.Equal(t, providerResponse.StatusCode, clientResponse.StatusCode)
	require.Equal(t, providerResponse.Headers, clientResponse.Headers)
	require.Equal(t, providerResponse.Body, clientResponse.Body)
}

func TestAlphaSearchInboundTransformer_RequiresModel(t *testing.T) {
	inbound := NewAlphaSearchInboundTransformer()
	_, err := inbound.TransformRequest(context.Background(), &httpclient.Request{
		Method: http.MethodPost,
		Body:   []byte(`{"commands":{}}`),
	})
	require.Error(t, err)
}

func TestAlphaSearchUsesHTTPWhenResponsesTransportIsWebSocket(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"computer_initialize_state"}`))
	}))
	defer server.Close()

	outbound, err := NewOutboundTransformer(Params{
		BaseURL:   server.URL,
		Transport: responses.TransportWebSocket,
		TokenProvider: staticTokenGetter{creds: &oauth.OAuthCredentials{
			AccessToken: "test-token",
			ExpiresAt:   time.Now().Add(time.Hour),
		}},
	})
	require.NoError(t, err)

	rawRequest := &httpclient.Request{Headers: make(http.Header), Body: []byte(`{"model":"gpt-5.6-sol","commands":{}}`)}
	providerRequest, err := outbound.TransformRequest(context.Background(), &llm.Request{
		Model:       "gpt-5.6-sol",
		RequestType: llm.RequestTypeAlphaSearch,
		APIFormat:   llm.APIFormatOpenAICodexAlphaSearch,
		RawRequest:  rawRequest,
	})
	require.NoError(t, err)

	executor := outbound.CustomizeExecutor(httpclient.NewHttpClientWithClient(server.Client()))
	response, err := executor.Do(context.Background(), providerRequest)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "/alpha/search", receivedPath)
}

func TestAlphaSearchSupportsCustomEndpointPath(t *testing.T) {
	outbound, err := NewOutboundTransformer(Params{
		BaseURL:         "https://example.com/backend-api/codex#",
		AlphaSearchPath: "/custom/search",
		TokenProvider: staticTokenGetter{creds: &oauth.OAuthCredentials{
			AccessToken: "test-token",
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "https://example.com/backend-api/codex/custom/search", outbound.alphaSearchURL())
}
