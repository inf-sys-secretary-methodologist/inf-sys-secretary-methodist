// ComposioConfig mirrors the JSON projection returned by
// GET /api/admin/composio/config. Only booleans surface — the
// Composio API key is a signing secret; entity ID and MCP config
// ID are opaque platform identifiers (per VAPID privacy precedent
// in admin/integrations).
export interface ComposioConfig {
  configured: boolean
  api_key_configured: boolean
  entity_id_set: boolean
  mcp_config_id_set: boolean
}
