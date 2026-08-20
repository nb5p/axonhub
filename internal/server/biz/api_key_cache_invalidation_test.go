package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/pkg/xcache/live"
)

type failingAPIKeyCacheNotifier struct{}

func (failingAPIKeyCacheNotifier) Watch() (<-chan live.CacheEvent[string], func()) {
	return make(chan live.CacheEvent[string]), func() {}
}

func (failingAPIKeyCacheNotifier) Notify(context.Context, live.CacheEvent[string]) error {
	return errors.New("watcher unavailable")
}

func TestAPIKeyService_InvalidateAPIKeyCachesInvalidatesLocalCacheSynchronously(t *testing.T) {
	cache := live.NewIndexedCache(live.IndexedOptions[string, *ent.APIKey]{
		Name:            "api_key_sync_invalidation_test",
		RefreshInterval: time.Hour,
		KeyFunc:         func(value *ent.APIKey) string { return buildAPIKeyCacheKey(value.Key) },
		LoadOneFunc: func(context.Context, string) (*ent.APIKey, error) {
			return nil, live.ErrKeyNotFound
		},
		LoadSinceFunc: func(context.Context, time.Time) ([]*ent.APIKey, time.Time, error) {
			return nil, time.Time{}, nil
		},
	})
	t.Cleanup(cache.Stop)

	apiKey := &ent.APIKey{Key: "ah-test-sync-invalidation"}
	cacheKey := buildAPIKeyCacheKey(apiKey.Key)
	cache.Set(cacheKey, apiKey)

	service := &APIKeyService{
		APIKeyCache:    cache,
		apiKeyNotifier: failingAPIKeyCacheNotifier{},
	}
	service.invalidateAPIKeyCaches(context.Background(), apiKey.Key)

	_, cached := cache.GetCached(cacheKey)
	require.False(t, cached, "the writer instance must not serve the old profile after the mutation returns")
}
