package codex

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	SessionHeader         = "Session_id"
	SessionHeaderHyphen   = "Session-Id"
	TurnMetadataHeader    = "X-Codex-Turn-Metadata"
	WindowIDHeader        = "X-Codex-Window-Id"
	ClientRequestIDHeader = "X-Client-Request-Id"
	BetaFeaturesHeader    = "X-Codex-Beta-Features"
	RemoteCompactionV2    = "remote_compaction_v2"
)

type TurnMetadata struct {
	SessionID string `json:"session_id"`
}

var PassthroughHeaders = []string{
	TurnMetadataHeader,
	WindowIDHeader,
	ClientRequestIDHeader,
	BetaFeaturesHeader,
}

func ExtractSessionIDFromTurnMetadata(raw string) string {
	if raw == "" {
		return ""
	}

	var payload TurnMetadata
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return ""
	}

	return strings.TrimSpace(payload.SessionID)
}

func GetSessionIDFromHeaders(headers http.Header) string {
	if headers == nil {
		return ""
	}

	sessionID := strings.TrimSpace(headers.Get(SessionHeader))
	if sessionID == "" {
		sessionID = strings.TrimSpace(headers.Get(SessionHeaderHyphen))
	}
	if sessionID != "" {
		return sessionID
	}

	return ExtractSessionIDFromTurnMetadata(strings.TrimSpace(headers.Get(TurnMetadataHeader)))
}

// HasBetaFeature reports whether Codex advertised an exact feature token.
// Feature names are comma-separated and case-sensitive on the upstream wire.
func HasBetaFeature(headers http.Header, feature string) bool {
	if headers == nil || feature == "" {
		return false
	}

	for _, value := range headers.Values(BetaFeaturesHeader) {
		for token := range strings.SplitSeq(value, ",") {
			if strings.TrimSpace(token) == feature {
				return true
			}
		}
	}

	return false
}
