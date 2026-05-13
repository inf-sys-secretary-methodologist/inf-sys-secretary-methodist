// Package integrations exposes a read-only admin view of the
// runtime configuration for the WebPush (VAPID) + n8n
// integrations. It does not change either runtime — webpush_service
// and the n8n event handler stay the single source of truth — it
// only reflects current config so admins can confirm wiring
// without reading server logs.
package integrations

import (
	"context"
	"os"
)

// VAPIDConfig is the JSON projection of the WebPush keypair
// configuration. The private key is never exposed — only its
// presence as a boolean — because it is a signing secret. The
// public key is safe to surface (the browser receives it via the
// /push/public-key endpoint anyway).
type VAPIDConfig struct {
	Configured bool   `json:"configured"`
	PublicKey  string `json:"public_key"`
	Subject    string `json:"subject"`
}

// N8NConfig mirrors the runtime n8n integration state. WebhookURL
// is non-secret (operational URL exposed in compose.yml) and is
// helpful for admins to verify they are pointed at the right
// instance.
type N8NConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
}

// Config is the combined response. Mirror к
// admin/sentry shape — a single read-only handler returns both
// projections so the frontend renders both status cards from one
// fetch.
type Config struct {
	VAPID VAPIDConfig `json:"vapid"`
	N8N   N8NConfig   `json:"n8n"`
}

// VAPIDProbe answers "is the VAPID keypair fully configured?"
// without exposing the private key. Both keys + subject must be
// non-empty for true. Tests substitute fakes for deterministic
// branch coverage.
type VAPIDProbe func() bool

// EnvVAPIDProbe is the production probe — both VAPID_PUBLIC_KEY
// AND VAPID_PRIVATE_KEY AND VAPID_SUBJECT must be set; missing
// any of the three means WebPush sender will fail at send time.
func EnvVAPIDProbe() bool {
	return os.Getenv("VAPID_PUBLIC_KEY") != "" &&
		os.Getenv("VAPID_PRIVATE_KEY") != "" &&
		os.Getenv("VAPID_SUBJECT") != ""
}

// AdminIntegrationsUseCase is the read-only admin view. Public
// VAPID key + subject + n8n enabled/webhook are sourced from the
// constructor-stored cfg snapshot; the configured boolean is
// computed by the injected probe.
type AdminIntegrationsUseCase struct {
	vapidProbe     VAPIDProbe
	vapidPublicKey string
	vapidSubject   string
	n8nEnabled     bool
	n8nWebhookURL  string
}

// NewAdminIntegrationsUseCase builds the view. vapidPublicKey +
// vapidSubject come from cfg.WebPush; n8nEnabled + n8nWebhookURL
// from cfg.N8N. Panics on nil probe so misconfigured DI fails at
// construction.
func NewAdminIntegrationsUseCase(
	vapidProbe VAPIDProbe,
	vapidPublicKey, vapidSubject string,
	n8nEnabled bool,
	n8nWebhookURL string,
) *AdminIntegrationsUseCase {
	if vapidProbe == nil {
		panic("integrations: nil VAPIDProbe")
	}
	return &AdminIntegrationsUseCase{
		vapidProbe:     vapidProbe,
		vapidPublicKey: vapidPublicKey,
		vapidSubject:   vapidSubject,
		n8nEnabled:     n8nEnabled,
		n8nWebhookURL:  n8nWebhookURL,
	}
}

// GetConfig returns the combined integrations snapshot. VAPID
// Configured is computed by the injected probe so tests can flip
// it deterministically; PublicKey / Subject / N8N fields come
// from the constructor-stored cfg snapshot.
func (uc *AdminIntegrationsUseCase) GetConfig(_ context.Context) Config {
	return Config{
		VAPID: VAPIDConfig{
			Configured: uc.vapidProbe(),
			PublicKey:  uc.vapidPublicKey,
			Subject:    uc.vapidSubject,
		},
		N8N: N8NConfig{
			Enabled:    uc.n8nEnabled,
			WebhookURL: uc.n8nWebhookURL,
		},
	}
}
