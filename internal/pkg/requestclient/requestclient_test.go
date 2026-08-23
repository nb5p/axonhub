package requestclient

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/llm"
)

func TestDetect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userAgent string
		want      llm.RequestClient
	}{
		{
			name:      "Codex Desktop",
			userAgent: "Codex Desktop/0.149.0-alpha.4.1 (Mac OS 26.5.1; arm64) unknown (Codex Desktop; 26.818.41509)",
			want:      llm.RequestClientCodex,
		},
		{name: "Codex CLI", userAgent: "codex-cli/0.1.0", want: llm.RequestClientCodex},
		{name: "legacy Codex CLI", userAgent: "codex_cli_rs/0.50.0 (macOS 14.0.0; arm64) Terminal", want: llm.RequestClientCodex},
		{name: "missing", userAgent: "", want: llm.RequestClientUnknown},
		{name: "other", userAgent: "curl/8.0", want: llm.RequestClientOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, DetectUserAgent(tt.userAgent))
		})
	}
}

func TestDetectReadsUserAgentHeader(t *testing.T) {
	t.Parallel()

	headers := http.Header{"User-Agent": {"Codex Desktop/0.149.0"}}
	require.Equal(t, llm.RequestClientCodex, Detect(headers))
}
