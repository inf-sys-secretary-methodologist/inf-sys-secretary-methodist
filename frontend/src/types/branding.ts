// BrandSettings mirrors the JSON projection returned by
// GET /api/admin/branding and GET /api/public/branding. The
// admin and public endpoints return the same shape — no field
// is sensitive (api keys / secrets never surface here).
export interface BrandSettings {
  app_name: string
  tagline: string
  logo_url: string
  favicon_url: string
  primary_color: string
  secondary_color: string
  updated_at: string
}

// UpdateBrandingRequest is the PUT body. Empty strings clear the
// corresponding field; backend validates each via domain entity
// invariants (hex color regex + URL scheme whitelist + length
// bounds) and maps validation failures to 422 с typed code.
export interface UpdateBrandingRequest {
  app_name: string
  tagline: string
  logo_url: string
  favicon_url: string
  primary_color: string
  secondary_color: string
}
