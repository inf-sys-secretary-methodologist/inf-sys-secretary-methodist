'use client'

// RED stub — GREEN ships the public-branding-aware client
// component (logo + app name + tagline + optional accent
// border). Stub renders only the testid so the role-guard /
// scaffold tests pass; behavioral tests fail.
interface BrandedHeaderProps {
  titleFallback: string
}

export function BrandedHeader(_props: BrandedHeaderProps) {
  return <div data-testid="branded-header" />
}
