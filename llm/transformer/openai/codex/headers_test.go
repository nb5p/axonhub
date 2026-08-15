package codex

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasBetaFeature(t *testing.T) {
	headers := make(http.Header)
	headers.Add(BetaFeaturesHeader, "responses_websockets_v2, remote_compaction_v2")
	headers.Add(BetaFeaturesHeader, "another_feature")

	require.True(t, HasBetaFeature(headers, RemoteCompactionV2))
	require.True(t, HasBetaFeature(headers, "another_feature"))
	require.False(t, HasBetaFeature(headers, "REMOTE_COMPACTION_V2"))
	require.False(t, HasBetaFeature(headers, "remote_compaction"))
	require.False(t, HasBetaFeature(nil, RemoteCompactionV2))
}
