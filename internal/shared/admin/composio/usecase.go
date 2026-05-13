// Package composio exposes a read-only admin view of the runtime
// Composio integration configuration. It does not touch the
// Composio client or any consumer — services in
// internal/modules/notifications/application/services stay the
// single source of truth — it only reflects current env-driven
// config so admins can confirm wiring without reading server logs.
package composio

import (
	"context"
	"os"
)

// Config is the JSON projection of the Composio runtime
// configuration. Only booleans surface — the API key is a signing
// secret, the entity ID and MCP config ID are opaque platform
// identifiers that do not serve admin observability beyond
// presence. Privacy-conservative default mirrors the VAPID private
// key precedent in admin/integrations.
type Config struct {
	Configured       bool `json:"configured"`
	APIKeyConfigured bool `json:"api_key_configured"`
	EntityIDSet      bool `json:"entity_id_set"`
	MCPConfigIDSet   bool `json:"mcp_config_id_set"`
}

// ProbeResult holds per-field configured booleans returned by the
// runtime probe. AllConfigured() is the aggregate predicate used
// to render the top-level status badge.
type ProbeResult struct {
	APIKeyConfigured bool
	EntityIDSet      bool
	MCPConfigIDSet   bool
}

// AllConfigured returns true when all three Composio env vars are
// non-empty. Composio services in notifications/ gate their own
// initialisation on APIKey AND EntityID (see main.go:339,347,1408);
// MCP integration additionally requires MCPConfigID. "Fully
// configured" therefore means all three.
func (r ProbeResult) AllConfigured() bool {
	return r.APIKeyConfigured && r.EntityIDSet && r.MCPConfigIDSet
}

// Probe answers "what is the runtime Composio env state?" Tests
// substitute fakes returning fixed ProbeResult values for
// deterministic branch coverage. Mirror к VAPIDProbe in
// admin/integrations, drift: struct-return rather than single
// bool, justified in plan ADR-2 (3 fields each need admin-visible
// status, no public values to surface from cfg snapshot).
type Probe func() ProbeResult

// EnvComposioProbe is the production probe — reads the three
// Composio env vars directly. Missing any of the three means the
// corresponding Composio capability is unwired (notifications
// services skip initialisation on empty APIKey or EntityID; MCP
// pipeline skips on empty MCPConfigID).
func EnvComposioProbe() ProbeResult {
	return ProbeResult{
		APIKeyConfigured: os.Getenv("COMPOSIO_API_KEY") != "",
		EntityIDSet:      os.Getenv("COMPOSIO_ENTITY_ID") != "",
		MCPConfigIDSet:   os.Getenv("COMPOSIO_MCP_CONFIG_ID") != "",
	}
}

// AdminComposioUseCase is the read-only admin view. The injected
// probe is the only data source — there is no cfg snapshot to
// surface because Composio fields are either secret (APIKey) or
// opaque platform identifiers (EntityID, MCPConfigID).
type AdminComposioUseCase struct {
	probe Probe
}

// NewAdminComposioUseCase builds the view. Panics on a nil probe
// so misconfigured DI fails at construction (mirror к
// admin/integrations + admin/sentry).
func NewAdminComposioUseCase(probe Probe) *AdminComposioUseCase {
	if probe == nil {
		panic("composio: nil Probe")
	}
	return &AdminComposioUseCase{probe: probe}
}

// GetConfig returns the Composio configuration projection. The
// aggregate Configured boolean is true only when all three env
// vars are non-empty; the per-field booleans surface so admins
// can see which specific field is missing when partial.
func (uc *AdminComposioUseCase) GetConfig(_ context.Context) Config {
	r := uc.probe()
	return Config{
		Configured:       r.AllConfigured(),
		APIKeyConfigured: r.APIKeyConfigured,
		EntityIDSet:      r.EntityIDSet,
		MCPConfigIDSet:   r.MCPConfigIDSet,
	}
}
