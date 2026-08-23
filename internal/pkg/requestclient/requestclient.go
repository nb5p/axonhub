// Package requestclient classifies inbound request clients from User-Agent.
package requestclient

import (
	"net/http"
	"strings"

	"github.com/looplj/axonhub/llm"
)

// Detect classifies a request User-Agent into a stable, intentionally small
// set of clients. Keep matching narrow: an accidental Codex classification
// would change a channel's eligible outbound API formats.
func Detect(headers http.Header) llm.RequestClient {
	return DetectUserAgent(headers.Get("User-Agent"))
}

// DetectUserAgent classifies one User-Agent value. It is exported for focused
// tests and for callers that only retained a User-Agent string.
func DetectUserAgent(userAgent string) llm.RequestClient {
	normalized := strings.ToLower(strings.TrimSpace(userAgent))
	if normalized == "" {
		return llm.RequestClientUnknown
	}

	if strings.HasPrefix(normalized, "codex desktop/") ||
		strings.HasPrefix(normalized, "codex-cli/") ||
		strings.HasPrefix(normalized, "codex_cli/") ||
		strings.HasPrefix(normalized, "codex_cli_rs/") {
		return llm.RequestClientCodex
	}

	return llm.RequestClientOther
}
