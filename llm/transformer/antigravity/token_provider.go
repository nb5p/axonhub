package antigravity

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/oauth"
)

// NewTokenProvider creates a new OAuth token provider for Antigravity.
func NewTokenProvider(params oauth.TokenProviderParams) *oauth.TokenProvider {
	params.OAuthUrls = DefaultTokenURLs
	if params.UserAgent == "" {
		params.UserAgent = GetUserAgent()
	}

	clientSecret := ClientSecret
	// Imported Gemini Code Assist credentials carry their own public OAuth
	// client ID. Supplying Antigravity's client secret to a different client
	// would make a standard refresh fail, so omit it in that case.
	if params.Credentials != nil && params.Credentials.ClientID != "" && params.Credentials.ClientID != ClientID {
		clientSecret = ""
	}
	params.ExchangeStrategy = &AntigravityExchangeStrategy{
		UserAgent:    params.UserAgent,
		ClientSecret: clientSecret,
	}

	return oauth.NewTokenProvider(params)
}

// DefaultTokenURLs are the Antigravity OAuth endpoints.
var DefaultTokenURLs = oauth.OAuthUrls{
	AuthorizeUrl: AuthorizeURL,
	TokenUrl:     TokenURL,
}

// AntigravityExchangeStrategy implements ExchangeStrategy for Google OAuth which requires ClientSecret.
type AntigravityExchangeStrategy struct {
	UserAgent    string
	ClientSecret string
}

// BuildExchangeRequest implements ExchangeStrategy.
func (s *AntigravityExchangeStrategy) BuildExchangeRequest(params oauth.ExchangeParams, tokenURL string) (*httpclient.Request, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", params.ClientID)
	if s.ClientSecret != "" {
		form.Set("client_secret", s.ClientSecret)
	}
	form.Set("code", params.Code)
	form.Set("redirect_uri", params.RedirectURI)
	form.Set("code_verifier", params.CodeVerifier)

	header := http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded"},
		"Accept":       []string{"application/json"},
	}
	if s.UserAgent != "" {
		header.Set("User-Agent", s.UserAgent)
	}

	return &httpclient.Request{
		Method:  http.MethodPost,
		URL:     tokenURL,
		Headers: header,
		Body:    []byte(form.Encode()),
	}, nil
}

// BuildRefreshRequest implements ExchangeStrategy.
func (s *AntigravityExchangeStrategy) BuildRefreshRequest(creds *oauth.OAuthCredentials, tokenURL string) (*httpclient.Request, error) {
	if creds == nil {
		return nil, errors.New("nil credentials")
	}

	if creds.RefreshToken == "" {
		return nil, errors.New("refresh_token is empty")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", creds.ClientID)
	if s.ClientSecret != "" {
		form.Set("client_secret", s.ClientSecret)
	}
	form.Set("refresh_token", creds.RefreshToken)

	header := http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded"},
		"Accept":       []string{"application/json"},
	}
	if s.UserAgent != "" {
		header.Set("User-Agent", s.UserAgent)
	}

	return &httpclient.Request{
		Method:  http.MethodPost,
		URL:     tokenURL,
		Headers: header,
		Body:    []byte(form.Encode()),
	}, nil
}
