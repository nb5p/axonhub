package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/sjson"

	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/streams"
	"github.com/looplj/axonhub/llm/transformer"
	"github.com/looplj/axonhub/llm/transformer/openai/responses"
	"github.com/looplj/axonhub/llm/transformer/shared"
)

type AlphaSearchInboundTransformer struct {
	responses *responses.InboundTransformer
}

type alphaSearchRequest struct {
	Model string `json:"model"`
}

type alphaSearchResponseMetadata struct {
	statusCode int
	headers    http.Header
	body       []byte
}

func NewAlphaSearchInboundTransformer() *AlphaSearchInboundTransformer {
	return &AlphaSearchInboundTransformer{responses: responses.NewInboundTransformer()}
}

func (t *AlphaSearchInboundTransformer) TransformRequest(_ context.Context, request *httpclient.Request) (*llm.Request, error) {
	if request == nil {
		return nil, fmt.Errorf("%w: request is nil", transformer.ErrInvalidRequest)
	}
	if len(request.Body) == 0 {
		return nil, fmt.Errorf("%w: request body is empty", transformer.ErrInvalidRequest)
	}

	var payload alphaSearchRequest
	if err := json.Unmarshal(request.Body, &payload); err != nil {
		return nil, fmt.Errorf("%w: failed to decode alpha search request: %w", transformer.ErrInvalidRequest, err)
	}
	if strings.TrimSpace(payload.Model) == "" {
		return nil, fmt.Errorf("%w: model is required", transformer.ErrInvalidRequest)
	}

	stream := false
	return &llm.Request{
		Model:       payload.Model,
		Stream:      &stream,
		RequestType: llm.RequestTypeAlphaSearch,
		APIFormat:   llm.APIFormatOpenAICodexAlphaSearch,
	}, nil
}

func (t *AlphaSearchInboundTransformer) TransformResponse(_ context.Context, response *llm.Response) (*httpclient.Response, error) {
	if response == nil || response.TransformerMetadata == nil {
		return nil, errors.New("alpha search response metadata is missing")
	}

	metadata, ok := response.TransformerMetadata[alphaSearchResponseMetadataKey].(*alphaSearchResponseMetadata)
	if !ok || metadata == nil {
		return nil, errors.New("alpha search response metadata is invalid")
	}

	return &httpclient.Response{
		StatusCode: metadata.statusCode,
		Headers:    metadata.headers.Clone(),
		Body:       append([]byte(nil), metadata.body...),
	}, nil
}

func (t *AlphaSearchInboundTransformer) TransformStream(context.Context, streams.Stream[*llm.Response]) (streams.Stream[*httpclient.StreamEvent], error) {
	return nil, errors.New("alpha search does not support streaming")
}

func (t *AlphaSearchInboundTransformer) TransformError(ctx context.Context, err error) *httpclient.Error {
	return t.responses.TransformError(ctx, err)
}

func (t *AlphaSearchInboundTransformer) AggregateStreamChunks(context.Context, []*httpclient.StreamEvent) ([]byte, llm.ResponseMeta, error) {
	return nil, llm.ResponseMeta{}, errors.New("alpha search does not support streaming")
}

func (t *OutboundTransformer) transformAlphaSearchRequest(ctx context.Context, llmReq *llm.Request) (*httpclient.Request, error) {
	if llmReq == nil || llmReq.RawRequest == nil {
		return nil, errors.New("alpha search raw request is missing")
	}

	creds, err := t.tokens.Get(ctx)
	if err != nil {
		return nil, err
	}
	if creds == nil || strings.TrimSpace(creds.AccessToken) == "" {
		return nil, errors.New("alpha search access token is missing")
	}

	body := append([]byte(nil), llmReq.RawRequest.Body...)
	if llmReq.Model != "" {
		body, err = sjson.SetBytes(body, "model", llmReq.Model)
		if err != nil {
			return nil, fmt.Errorf("set alpha search model: %w", err)
		}
	}

	headers := make(http.Header)
	headers.Set("Accept", "application/json")
	headers.Set("Content-Type", "application/json")

	rawHeaders := llmReq.RawRequest.Headers
	sessionID := GetSessionIDFromHeaders(rawHeaders)
	// Normalize the underscore variant before MergeInboundRequest runs. Otherwise
	// the original Session_id header is merged back alongside Session-Id.
	rawHeaders.Del(SessionHeader)

	if originator := rawHeaders.Get("Originator"); originator != "" {
		headers.Set("Originator", originator)
	} else {
		headers.Set("Originator", AxonHubOriginator)
	}
	if userAgent := rawHeaders.Get("User-Agent"); userAgent != "" {
		headers.Set("User-Agent", userAgent)
	}
	for _, header := range PassthroughHeaders {
		if value := rawHeaders.Get(header); value != "" {
			headers.Set(header, value)
		}
	}

	if sessionID == "" {
		if sharedSessionID, ok := shared.GetSessionID(ctx); ok {
			sessionID = sharedSessionID
		}
	}
	if sessionID != "" {
		headers.Set(SessionHeaderHyphen, sessionID)
		headers.Set("Conversation_id", sessionID)
	}
	if accountID := ExtractChatGPTAccountIDFromJWT(creds.AccessToken); accountID != "" {
		headers.Set("Chatgpt-Account-Id", accountID)
	}
	if version := rawHeaders.Get("Version"); version != "" {
		headers.Set("Version", version)
	} else {
		headers.Set("Version", codexDefaultVersion)
	}

	return &httpclient.Request{
		Method:      http.MethodPost,
		URL:         t.alphaSearchURL(),
		Headers:     headers,
		Body:        body,
		Auth:        &httpclient.AuthConfig{Type: httpclient.AuthTypeBearer, APIKey: creds.AccessToken},
		RequestType: llm.RequestTypeAlphaSearch.String(),
		APIFormat:   llm.APIFormatOpenAICodexAlphaSearch.String(),
	}, nil
}

func (t *OutboundTransformer) alphaSearchURL() string {
	baseURL := strings.TrimRight(strings.TrimSpace(t.baseURL), "/#")
	baseURL = strings.TrimSuffix(baseURL, "/responses")
	path := strings.TrimSpace(t.alphaSearchPath)
	if path == "" {
		path = "/alpha/search"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	parsed, err := url.Parse(baseURL)
	if err == nil {
		switch parsed.Scheme {
		case "ws":
			parsed.Scheme = "http"
		case "wss":
			parsed.Scheme = "https"
		}
		baseURL = parsed.String()
	}

	return strings.TrimRight(baseURL, "/") + path
}

func transformAlphaSearchResponse(httpResp *httpclient.Response) (*llm.Response, error) {
	if httpResp == nil {
		return nil, errors.New("http response is nil")
	}
	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return nil, &httpclient.Error{StatusCode: httpResp.StatusCode, Body: httpResp.Body}
	}

	return &llm.Response{
		Object:      "codex.alpha_search",
		RequestType: llm.RequestTypeAlphaSearch,
		APIFormat:   llm.APIFormatOpenAICodexAlphaSearch,
		TransformerMetadata: map[string]any{
			alphaSearchResponseMetadataKey: &alphaSearchResponseMetadata{
				statusCode: httpResp.StatusCode,
				headers:    httpResp.Headers.Clone(),
				body:       append([]byte(nil), httpResp.Body...),
			},
		},
	}, nil
}
