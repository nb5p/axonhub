package orchestrator

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/transformer/openai/codex"
)

func TestCodexTurnStateOriginStore(t *testing.T) {
	store := newCodexTurnStateOriginStore()

	_, ok := store.mintedBy("seed-1")
	require.False(t, ok)

	store.note("seed-1", "key:credA", codexTurnStateProvenanceTTL)
	identity, ok := store.mintedBy("seed-1")
	require.True(t, ok)
	require.Equal(t, "key:credA", identity)

	// Overwrite with a new minter.
	store.note("seed-1", "key:credB", codexTurnStateProvenanceTTL)
	identity, ok = store.mintedBy("seed-1")
	require.True(t, ok)
	require.Equal(t, "key:credB", identity)
}

func TestCodexTurnStateOriginStoreExpiry(t *testing.T) {
	store := newCodexTurnStateOriginStore()
	store.note("seed-expired", "key:credA", time.Millisecond)

	time.Sleep(5 * time.Millisecond)

	_, ok := store.mintedBy("seed-expired")
	require.False(t, ok)

	// The lazy read drops the record; a sweep finds nothing to do.
	store.sweep()
	_, ok = store.mintedBy("seed-expired")
	require.False(t, ok)
}

func TestCodexTurnStateSeed(t *testing.T) {
	apiKey := &ent.APIKey{ID: 42}

	cases := []struct {
		name    string
		headers http.Header
		want    string
	}{
		{
			name:    "hyphen session header",
			headers: http.Header{codex.SessionHeaderHyphen: []string{"session-abc"}},
			want:    "42\x00session-abc",
		},
		{
			name:    "underscore session header",
			headers: http.Header{codex.SessionHeader: []string{"session-abc"}},
			want:    "42\x00session-abc",
		},
		{
			name:    "turn metadata session",
			headers: http.Header{codex.TurnMetadataHeader: []string{`{"session_id":"session-abc"}`}},
			want:    "42\x00session-abc",
		},
		{
			name:    "no session identifier",
			headers: http.Header{},
			want:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, codexTurnStateSeed(apiKey, tc.headers))
		})
	}

	require.Equal(t, "", codexTurnStateSeed(nil, http.Header{codex.SessionHeaderHyphen: []string{"session-abc"}}))
}

func TestCodexTurnStateGuard_SameCredentialKeepsEcho(t *testing.T) {
	outbound, state := newCodexTurnStateGuardFixture(t, "key:credA")
	ctx := contexts.EnsureContainer(context.Background())
	contexts.WithChannelAPIKey(ctx, "credA")

	request := &httpclient.Request{
		Headers: http.Header{codex.TurnStateHeader: []string{"blob-echo"}},
	}

	middleware := withCodexTurnStateGuard(outbound, newCodexTurnStateOriginStore())
	got, err := middleware.OnOutboundRawRequest(ctx, request)
	require.NoError(t, err)
	require.Equal(t, "blob-echo", got.Headers.Get(codex.TurnStateHeader))
	require.Equal(t, "key:credA", state.CodexTurnStateMintedIdentity)
}

func TestCodexTurnStateGuard_CrossCredentialStripsEcho(t *testing.T) {
	store := newCodexTurnStateOriginStore()
	store.note("42\x00session-abc", "key:credA", codexTurnStateProvenanceTTL)

	outbound, state := newCodexTurnStateGuardFixture(t, "key:credB")
	ctx := contexts.EnsureContainer(context.Background())
	contexts.WithChannelAPIKey(ctx, "credB")

	request := &httpclient.Request{
		Headers: http.Header{codex.TurnStateHeader: []string{"blob-echo"}},
	}

	middleware := withCodexTurnStateGuard(outbound, store)
	got, err := middleware.OnOutboundRawRequest(ctx, request)
	require.NoError(t, err)
	require.Empty(t, got.Headers.Get(codex.TurnStateHeader))
	require.Equal(t, "key:credB", state.CodexTurnStateMintedIdentity)
}

func TestCodexTurnStateGuard_UnknownProvenanceKeepsEcho(t *testing.T) {
	outbound, state := newCodexTurnStateGuardFixture(t, "key:credA")
	ctx := contexts.EnsureContainer(context.Background())
	contexts.WithChannelAPIKey(ctx, "credA")

	request := &httpclient.Request{
		Headers: http.Header{codex.TurnStateHeader: []string{"blob-echo"}},
	}

	middleware := withCodexTurnStateGuard(outbound, newCodexTurnStateOriginStore())
	got, err := middleware.OnOutboundRawRequest(ctx, request)
	require.NoError(t, err)
	require.Equal(t, "blob-echo", got.Headers.Get(codex.TurnStateHeader))
	require.Equal(t, "key:credA", state.CodexTurnStateMintedIdentity)
}

func TestCodexTurnStateGuard_NoSeedKeepsEchoUntracked(t *testing.T) {
	outbound, state := newCodexTurnStateGuardFixture(t, "key:credA")
	state.CodexTurnStateSeed = ""

	ctx := contexts.EnsureContainer(context.Background())
	contexts.WithChannelAPIKey(ctx, "credA")

	request := &httpclient.Request{
		Headers: http.Header{codex.TurnStateHeader: []string{"blob-echo"}},
	}

	middleware := withCodexTurnStateGuard(outbound, newCodexTurnStateOriginStore())
	got, err := middleware.OnOutboundRawRequest(ctx, request)
	require.NoError(t, err)
	require.Equal(t, "blob-echo", got.Headers.Get(codex.TurnStateHeader))
	require.Empty(t, state.CodexTurnStateMintedIdentity)
}

func TestCodexTurnStateGuard_OAuthAccountIdentity(t *testing.T) {
	store := newCodexTurnStateOriginStore()
	store.note("42\x00session-abc", "oauth:acct-1", codexTurnStateProvenanceTTL)

	channel := &biz.Channel{Channel: &ent.Channel{
		ID: 7,
		Credentials: objects.ChannelCredentials{
			OAuth: &objects.OAuthCredentials{AccessToken: "stub"},
		},
	}}
	outbound, state := newCodexTurnStateGuardFixtureWithChannel(t, channel)

	// Same account: keep the echo.
	request := &httpclient.Request{
		Headers: http.Header{
			codex.TurnStateHeader: []string{"blob-echo"},
			"Chatgpt-Account-Id":  []string{"acct-1"},
		},
	}
	got, err := withCodexTurnStateGuard(outbound, store).OnOutboundRawRequest(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, "blob-echo", got.Headers.Get(codex.TurnStateHeader))
	require.Equal(t, "oauth:acct-1", state.CodexTurnStateMintedIdentity)

	// Different account: strip the echo even though the context has no key.
	state.CodexTurnStateMintedIdentity = ""
	request = &httpclient.Request{
		Headers: http.Header{
			codex.TurnStateHeader: []string{"blob-echo"},
			"Chatgpt-Account-Id":  []string{"acct-2"},
		},
	}
	got, err = withCodexTurnStateGuard(outbound, store).OnOutboundRawRequest(context.Background(), request)
	require.NoError(t, err)
	require.Empty(t, got.Headers.Get(codex.TurnStateHeader))
	require.Equal(t, "oauth:acct-2", state.CodexTurnStateMintedIdentity)
}

func TestCodexTurnStateGuard_OAuthWithoutAccountHeaderFallsBackToChannel(t *testing.T) {
	store := newCodexTurnStateOriginStore()
	store.note("42\x00session-abc", "oauth:channel:7", codexTurnStateProvenanceTTL)

	channel := &biz.Channel{Channel: &ent.Channel{
		ID: 7,
		Credentials: objects.ChannelCredentials{
			OAuth: &objects.OAuthCredentials{AccessToken: "stub"},
		},
	}}
	outbound, state := newCodexTurnStateGuardFixtureWithChannel(t, channel)

	request := &httpclient.Request{
		Headers: http.Header{codex.TurnStateHeader: []string{"blob-echo"}},
	}
	got, err := withCodexTurnStateGuard(outbound, store).OnOutboundRawRequest(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, "blob-echo", got.Headers.Get(codex.TurnStateHeader))
	require.Equal(t, "oauth:channel:7", state.CodexTurnStateMintedIdentity)
}

func newCodexTurnStateGuardFixture(t *testing.T, key string) (*PersistentOutboundTransformer, *PersistenceState) {
	t.Helper()

	channel := &biz.Channel{Channel: &ent.Channel{
		ID:          1,
		Credentials: objects.ChannelCredentials{APIKey: key},
	}}

	return newCodexTurnStateGuardFixtureWithChannel(t, channel)
}

func newCodexTurnStateGuardFixtureWithChannel(t *testing.T, channel *biz.Channel) (*PersistentOutboundTransformer, *PersistenceState) {
	t.Helper()

	state := &PersistenceState{
		APIKey:             &ent.APIKey{ID: 42},
		CodexTurnStateSeed: "42\x00session-abc",
		CurrentCandidate: &ChannelModelsCandidate{
			Channel: channel,
		},
	}

	return &PersistentOutboundTransformer{state: state}, state
}
