// IntegrationsConfig mirrors the JSON projection returned by
// GET /api/admin/integrations/config. VAPID private key never
// appears here — only the Configured boolean — because raw
// private keys are signing secrets even on an admin-gated
// endpoint.
export interface IntegrationsConfig {
  vapid: VAPIDConfig
  n8n: N8NConfig
}

export interface VAPIDConfig {
  configured: boolean
  public_key: string
  subject: string
}

export interface N8NConfig {
  enabled: boolean
  webhook_url: string
}
