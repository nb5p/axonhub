package orchestrator

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/pipeline"
	"github.com/looplj/axonhub/llm/transformer/openai/codex"
)

// codexTurnStateProvenanceTTL bounds how long a turn-state provenance record
// stays fresh. The blob is minted by the upstream under an outbound identity
// (installation/session/thread), so a stale record only ever causes a
// conservative strip of the echo on a later request, never a wrong keep.
const codexTurnStateProvenanceTTL = 24 * time.Hour

// codexTurnStateOrigin records which credential minted the last
// X-Codex-Turn-State blob for a downstream session.
type codexTurnStateOrigin struct {
	identity  string
	expiresAt time.Time
}

// codexTurnStateOriginStore keeps the provenance map process-wide. It is
// intentionally shared across requests: the client echoes the blob on the next
// turn, so the guard needs the minting credential of a previous request.
//
// Keys have no natural upper bound (one entry per downstream session), so the
// store sweeps expired entries on the read side and opportunistically once
// every 256 writes.
type codexTurnStateOriginStore struct {
	mu      sync.Mutex
	origins map[string]codexTurnStateOrigin
	writes  uint64
}

func newCodexTurnStateOriginStore() *codexTurnStateOriginStore {
	return &codexTurnStateOriginStore{origins: make(map[string]codexTurnStateOrigin)}
}

// note records that the given credential minted the blob for seed.
func (s *codexTurnStateOriginStore) note(seed, identity string, ttl time.Duration) {
	if s == nil || seed == "" || identity == "" || ttl <= 0 {
		return
	}

	s.mu.Lock()
	s.origins[seed] = codexTurnStateOrigin{
		identity:  identity,
		expiresAt: time.Now().Add(ttl),
	}
	s.writes++
	doSweep := s.writes%256 == 0
	s.mu.Unlock()

	if doSweep {
		s.sweep()
	}
}

// mintedBy returns the credential identity recorded for seed, lazily dropping
// expired records.
func (s *codexTurnStateOriginStore) mintedBy(seed string) (string, bool) {
	if s == nil || seed == "" {
		return "", false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	origin, ok := s.origins[seed]
	if !ok {
		return "", false
	}
	if !origin.expiresAt.IsZero() && time.Now().After(origin.expiresAt) {
		delete(s.origins, seed)
		return "", false
	}

	return origin.identity, true
}

// sweep removes all expired records. It is called from note every 256 writes
// and is safe to call concurrently.
func (s *codexTurnStateOriginStore) sweep() {
	if s == nil {
		return
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for seed, origin := range s.origins {
		if origin.expiresAt.IsZero() || now.After(origin.expiresAt) {
			delete(s.origins, seed)
		}
	}
}

// codexTurnStateSeed returns the provenance key for a downstream session: the
// downstream API key ID plus the client session identifier observed at ingress.
// Without a session identifier there is no way to correlate the echo with its
// minting response, so the guard stays out of the way (current pass-through
// behavior).
func codexTurnStateSeed(apiKey *ent.APIKey, headers http.Header) string {
	if apiKey == nil {
		return ""
	}

	sessionID := codex.GetSessionIDFromHeaders(headers)
	if sessionID == "" {
		return ""
	}

	return strconv.FormatInt(int64(apiKey.ID), 10) + "\x00" + sessionID
}

// codexTurnStateAttemptIdentity identifies the credential an outbound attempt
// authenticates as. The identity is only used to compare minters across
// attempts, never logged verbatim (API keys may be sensitive).
//
// OAuth channels are checked before the context value because the context
// container is shared across attempts: a previous API-key attempt may have
// left a stale key while the current attempt authenticates with OAuth.
func codexTurnStateAttemptIdentity(ctx context.Context, channel *biz.Channel, request *httpclient.Request) string {
	if request == nil || request.Headers == nil {
		return ""
	}

	if channel != nil && channel.Credentials.IsOAuth() {
		if accountID := request.Headers.Get("Chatgpt-Account-Id"); accountID != "" {
			return "oauth:" + accountID
		}
		return fmt.Sprintf("oauth:channel:%d", channel.ID)
	}

	if apiKey, ok := contexts.GetChannelAPIKey(ctx); ok && apiKey != "" {
		return "key:" + apiKey
	}

	if channel != nil {
		return fmt.Sprintf("channel:%d", channel.ID)
	}

	return ""
}

// withCodexTurnStateGuard keeps the Codex round-trip consistent across
// channel/credential failover. The upstream mints X-Codex-Turn-State under the
// outbound identity; the client echoes it back on the next turn. Replaying a
// blob minted by account A into a request routed to account B is a
// proxy-only contradiction that real Codex can never produce, so the guard
// strips it. Same-credential echoes and unknown provenance pass through
// untouched.
//
// The middleware also records the current attempt's credential identity on the
// state; the orchestrator commits provenance only after the final upstream
// response actually carries the blob (see Process).
func withCodexTurnStateGuard(outbound *PersistentOutboundTransformer, store *codexTurnStateOriginStore) pipeline.Middleware {
	return pipeline.OnRawRequest("codex-turn-state-guard", func(ctx context.Context, request *httpclient.Request) (*httpclient.Request, error) {
		if outbound == nil || outbound.state == nil || store == nil || outbound.state.CodexTurnStateSeed == "" {
			return request, nil
		}
		if request == nil || request.Headers == nil || strings.TrimSpace(request.Headers.Get(codex.TurnStateHeader)) == "" {
			return request, nil
		}

		channel := outbound.GetCurrentChannel()
		identity := codexTurnStateAttemptIdentity(ctx, channel, request)
		if identity == "" {
			return request, nil
		}

		outbound.state.CodexTurnStateMintedIdentity = identity

		if mintedBy, ok := store.mintedBy(outbound.state.CodexTurnStateSeed); ok && mintedBy != identity {
			request.Headers.Del(codex.TurnStateHeader)
		}

		return request, nil
	})
}
