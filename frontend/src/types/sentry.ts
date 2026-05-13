// SentryConfig mirrors the JSON projection returned by
// GET /api/admin/sentry/config. DSN value is never returned — only its
// presence as boolean — because the raw DSN is a secret even on an
// admin-gated endpoint.
export interface SentryConfig {
  dsn_configured: boolean
  environment: string
  release: string
  traces_sample_rate: number
  tracing_enabled: boolean
}
