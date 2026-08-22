package biz

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/enttest"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/xcache"
	"github.com/looplj/axonhub/llm/httpclient"
)

func newTestSystemServiceWithWebhookConfig(t *testing.T, client *ent.Client, cfg WebhookNotifierConfig) *SystemService {
	t.Helper()

	service := &SystemService{
		AbstractService: &AbstractService{
			db: client,
		},
		Cache: xcache.NewFromConfig[ent.System](xcache.Config{Mode: xcache.ModeMemory}),
	}

	ctx := ent.NewContext(context.Background(), client)
	ctx = authz.WithTestBypass(ctx)
	require.NoError(t, service.SetWebhookNotifierConfig(ctx, &cfg))

	return service
}

func TestWebhookNotifier_NotifyChannelAutoDisabled(t *testing.T) {
	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=0")
	defer client.Close()

	var (
		receivedBody   string
		receivedHeader string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		receivedBody = string(body)
		receivedHeader = r.Header.Get("X-Axonhub-Event")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := WebhookNotifierConfig{
		Targets: []WebhookTarget{
			{
				Name:      "default",
				Enabled:   true,
				URL:       server.URL,
				TimeoutMs: 1000,
				Headers: []objects.HeaderEntry{
					{Key: "X-AxonHub-Event", Value: "{{.Event}}"},
				},
				Body: `{"event":"{{.Event}}","channel":"{{.Channel.Name}}","status_code":{{.Trigger.StatusCode}},"threshold":{{.Trigger.Threshold}},"actual_count":{{.Trigger.ActualCount}}}`,
			},
		},
		Subscriptions: []WebhookSubscription{
			{Event: EventChannelAutoDisabled, TargetNames: []string{"default"}},
		},
	}

	systemService := newTestSystemServiceWithWebhookConfig(t, client, cfg)
	notifier := NewWebhookNotifier(systemService, httpclient.NewHttpClient())

	notifier.NotifyChannelAutoDisabled(context.Background(), ChannelAutoDisabledEvent{
		ChannelID:       1,
		ChannelName:     "primary",
		ChannelProvider: "openai",
		ChannelBaseURL:  "https://api.openai.com",
		ChannelStatus:   "disabled",
		StatusCode:      429,
		Threshold:       3,
		ActualCount:     3,
		Reason:          "quota exhausted",
		OccurredAt:      time.Unix(1712812800, 0),
	})

	require.Equal(t, EventChannelAutoDisabled, receivedHeader)
	require.Contains(t, receivedBody, `"event":"channel.auto_disabled"`)
	require.Contains(t, receivedBody, `"channel":"primary"`)
	require.Contains(t, receivedBody, `"status_code":429`)
	require.Contains(t, receivedBody, `"threshold":3`)
	require.Contains(t, receivedBody, `"actual_count":3`)
}

func TestWebhookNotifier_SkipWhenTemplateInvalid(t *testing.T) {
	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=0")
	defer client.Close()

	called := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := WebhookNotifierConfig{
		Targets: []WebhookTarget{
			{
				Name:    "default",
				Enabled: true,
				URL:     server.URL,
				Body:    `{"event":"{{if .Event}}"}`,
			},
		},
		Subscriptions: []WebhookSubscription{
			{Event: EventChannelAutoDisabled, TargetNames: []string{"default"}},
		},
	}

	systemService := newTestSystemServiceWithWebhookConfig(t, client, cfg)
	notifier := NewWebhookNotifier(systemService, httpclient.NewHttpClient())
	notifier.NotifyChannelAutoDisabled(context.Background(), ChannelAutoDisabledEvent{OccurredAt: time.Now()})

	require.False(t, called)
}

func TestNormalizeWebhookNotifierConfig_InitializesDefaults(t *testing.T) {
	cfg := WebhookNotifierConfig{}

	normalizeWebhookNotifierConfig(&cfg)

	require.NotNil(t, cfg.Targets)
	require.NotNil(t, cfg.Subscriptions)
}

func TestWebhookNotifier_SelectTargetsSkipsInvalidTargets(t *testing.T) {
	notifier := &WebhookNotifier{}
	targets := notifier.selectTargets(WebhookNotifierConfig{
		Targets: []WebhookTarget{
			{Name: "a", Enabled: true, URL: "https://example.com"},
			{Name: "b", Enabled: false, URL: "https://example.com"},
			{Name: "c", Enabled: true, URL: ""},
		},
		Subscriptions: []WebhookSubscription{
			{Event: EventChannelAutoDisabled, TargetNames: []string{"a", "b", "c", "missing"}},
		},
	}, EventChannelAutoDisabled)

	require.Len(t, targets, 1)
	require.Equal(t, "a", targets[0].Name)
}

func TestRenderWebhookTemplate_NoTemplate(t *testing.T) {
	result, err := renderWebhookTemplate("plain text", WebhookRenderContext{})
	require.NoError(t, err)
	require.Equal(t, "plain text", result)
}

func TestWebhookNotifier_RecordsMaskedFailedDelivery(t *testing.T) {
	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=0")
	defer client.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := WebhookNotifierConfig{
		Targets: []WebhookTarget{{
			Name:    "masked",
			Enabled: true,
			URL:     server.URL,
			Headers: []objects.HeaderEntry{{Key: "Authorization", Value: "Bearer webhook-test-token"}},
			Body:    `{"token":"payload-secret"}`,
		}},
		Subscriptions: []WebhookSubscription{{Event: EventChannelAutoDisabled, TargetNames: []string{"masked"}}},
	}

	systemService := newTestSystemServiceWithWebhookConfig(t, client, cfg)
	notifier := NewWebhookNotifier(systemService, httpclient.NewHttpClient())
	notifier.NotifyChannelAutoDisabled(context.Background(), ChannelAutoDisabledEvent{OccurredAt: time.Now()})

	history, err := notifier.DeliveryHistory(authz.WithTestBypass(ent.NewContext(context.Background(), client)), 10)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, "failed", history[0].Status)
	require.Equal(t, http.StatusBadRequest, history[0].ResponseStatus)
	require.NotContains(t, history[0].RequestHeaders[0].Value, "webhook-test-token")
	require.NotContains(t, history[0].RequestBody, "payload-secret")
	require.Contains(t, history[0].RequestBody, "****cret")
}

func TestWebhookNotifier_BarkTargetUsesPushAPIAndMasksDeviceKey(t *testing.T) {
	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=0")
	defer client.Close()

	received := make(chan map[string]string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/push", r.URL.Path)
		payload := map[string]string{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		received <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := WebhookNotifierConfig{
		Targets: []WebhookTarget{{
			Name:          "bark",
			Enabled:       true,
			Type:          WebhookTargetTypeBark,
			URL:           server.URL,
			BarkDeviceKey: "bark-device-key",
			BarkTitle:     "{{.Event}}",
			Body:          "remaining {{.Quota.Remaining}} {{.Quota.Unit}}",
			BarkLevel:     "active",
			BarkGroup:     "axonhub",
		}},
		Subscriptions: []WebhookSubscription{{Event: EventQuotaBalanceExhausted, TargetNames: []string{"bark"}}},
	}

	systemService := newTestSystemServiceWithWebhookConfig(t, client, cfg)
	notifier := NewWebhookNotifier(systemService, httpclient.NewHttpClient())
	notifier.NotifyQuotaBalanceExhaustedAsync(context.Background(), QuotaBalanceExhaustedEvent{
		Remaining:  0,
		Unit:       "USD",
		OccurredAt: time.Now(),
	})

	select {
	case payload := <-received:
		require.Equal(t, "bark-device-key", payload["device_key"])
		require.Equal(t, EventQuotaBalanceExhausted, payload["title"])
		require.Equal(t, "remaining 0 USD", payload["body"])
		require.Equal(t, "active", payload["level"])
		require.Equal(t, "axonhub", payload["group"])
	case <-time.After(time.Second):
		t.Fatal("Bark target did not receive a request")
	}

	require.Eventually(t, func() bool {
		history, err := notifier.DeliveryHistory(authz.WithTestBypass(ent.NewContext(context.Background(), client)), 10)
		return err == nil && len(history) == 1
	}, time.Second, 10*time.Millisecond)

	history, err := notifier.DeliveryHistory(authz.WithTestBypass(ent.NewContext(context.Background(), client)), 10)
	require.NoError(t, err)
	require.Equal(t, "success", history[0].Status)
	require.NotContains(t, history[0].RequestBody, "bark-device-key")
	require.Contains(t, history[0].RequestBody, "****-key")
}
