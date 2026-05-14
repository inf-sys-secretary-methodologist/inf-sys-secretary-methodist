// Package entities holds the BrandSettings aggregate — the editable
// system brand (app name + logo + favicon + accent colors + tagline)
// shown on the login page and across the admin chrome. Greenfield in
// v0.136.0 — no prior brand seam existed.
package entities

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ErrInvalidAppName — app name must be non-empty and ≤ MaxAppNameLen.
// Mapped к HTTP 422 by the admin handler.
var ErrInvalidAppName = errors.New("branding: invalid app name")

// ErrInvalidTagline — tagline must be ≤ MaxTaglineLen (empty allowed).
var ErrInvalidTagline = errors.New("branding: invalid tagline")

// ErrInvalidColor — color must match #RRGGBB or #RGB hex format
// (empty allowed — optional field).
var ErrInvalidColor = errors.New("branding: invalid hex color")

// ErrInvalidURL — URL must be parseable AND scheme ∈ {http, https}
// (empty allowed — optional field). http/https whitelist prevents
// javascript:/data:/file: schemes from sneaking into the login page
// renderer (defense-in-depth before React).
var ErrInvalidURL = errors.New("branding: invalid URL")

const (
	// MaxAppNameLen — bound chosen to fit comfortably in browser tab
	// titles and the login page header on mobile breakpoints.
	MaxAppNameLen = 100
	// MaxTaglineLen — bound chosen for a one-line tagline under the
	// header. Markdown / multi-line tagline is out of scope.
	MaxTaglineLen = 200
)

var hexColorRE = regexp.MustCompile(`^#([0-9a-fA-F]{6}|[0-9a-fA-F]{3})$`)

// BrandSettings is the singleton aggregate root for the system brand.
// Persisted as a single row (id=1) in the brand_settings table —
// migration 037 enforces the singleton invariant с CHECK (id=1).
// All fields except AppName are optional; the admin handler PUTs the
// full snapshot every time (no PATCH semantics).
type BrandSettings struct {
	appName        string
	tagline        string
	logoURL        string
	faviconURL     string
	primaryColor   string
	secondaryColor string
	updatedAt      time.Time
}

// NewBrandSettings builds a BrandSettings instance validating every
// invariant in one shot. Returns the first violation as a typed
// sentinel for errors.Is matching. UpdatedAt is set to the provided
// now value so tests can pin deterministic timestamps.
func NewBrandSettings(
	appName, tagline, logoURL, faviconURL, primaryColor, secondaryColor string,
	now time.Time,
) (*BrandSettings, error) {
	if err := validateAppName(appName); err != nil {
		return nil, err
	}
	if err := validateTagline(tagline); err != nil {
		return nil, err
	}
	if err := validateURL(logoURL); err != nil {
		return nil, err
	}
	if err := validateURL(faviconURL); err != nil {
		return nil, err
	}
	if err := validateColor(primaryColor); err != nil {
		return nil, err
	}
	if err := validateColor(secondaryColor); err != nil {
		return nil, err
	}
	return &BrandSettings{
		appName:        appName,
		tagline:        tagline,
		logoURL:        logoURL,
		faviconURL:     faviconURL,
		primaryColor:   primaryColor,
		secondaryColor: secondaryColor,
		updatedAt:      now,
	}, nil
}

// RehydrateBrandSettings reconstructs a BrandSettings from persisted
// state — bypasses constructor validation under the assumption that
// stored data was previously validated. Repository-only entry point.
func RehydrateBrandSettings(
	appName, tagline, logoURL, faviconURL, primaryColor, secondaryColor string,
	updatedAt time.Time,
) *BrandSettings {
	return &BrandSettings{
		appName:        appName,
		tagline:        tagline,
		logoURL:        logoURL,
		faviconURL:     faviconURL,
		primaryColor:   primaryColor,
		secondaryColor: secondaryColor,
		updatedAt:      updatedAt,
	}
}

// AppName returns the configured system name shown in the login
// header, browser tab, and admin chrome.
func (b *BrandSettings) AppName() string { return b.appName }

// Tagline returns the optional one-liner shown under the app name
// on the login page.
func (b *BrandSettings) Tagline() string { return b.tagline }

// LogoURL returns the optional logo URL rendered on the login page.
func (b *BrandSettings) LogoURL() string { return b.logoURL }

// FaviconURL returns the optional favicon URL surfaced via the
// public branding endpoint.
func (b *BrandSettings) FaviconURL() string { return b.faviconURL }

// PrimaryColor returns the optional primary accent color (hex).
func (b *BrandSettings) PrimaryColor() string { return b.primaryColor }

// SecondaryColor returns the optional secondary accent color (hex).
func (b *BrandSettings) SecondaryColor() string { return b.secondaryColor }

// UpdatedAt returns the timestamp of the most recent successful
// mutation — either the constructor's now or the latest Update*.
func (b *BrandSettings) UpdatedAt() time.Time { return b.updatedAt }

// UpdateAppName replaces the app name in-place; touches updatedAt
// only on successful validation. The previous value is preserved
// when validation fails so callers can render the rejected form
// without losing state.
func (b *BrandSettings) UpdateAppName(name string, now time.Time) error {
	if err := validateAppName(name); err != nil {
		return err
	}
	b.appName = name
	b.updatedAt = now
	return nil
}

// UpdateTagline replaces the tagline; empty value clears it.
func (b *BrandSettings) UpdateTagline(tagline string, now time.Time) error {
	if err := validateTagline(tagline); err != nil {
		return err
	}
	b.tagline = tagline
	b.updatedAt = now
	return nil
}

// UpdateLogoURL replaces the logo URL; empty clears it.
func (b *BrandSettings) UpdateLogoURL(u string, now time.Time) error {
	if err := validateURL(u); err != nil {
		return err
	}
	b.logoURL = u
	b.updatedAt = now
	return nil
}

// UpdateFaviconURL replaces the favicon URL; empty clears it.
func (b *BrandSettings) UpdateFaviconURL(u string, now time.Time) error {
	if err := validateURL(u); err != nil {
		return err
	}
	b.faviconURL = u
	b.updatedAt = now
	return nil
}

// UpdatePrimaryColor replaces the primary accent color; empty
// clears it.
func (b *BrandSettings) UpdatePrimaryColor(c string, now time.Time) error {
	if err := validateColor(c); err != nil {
		return err
	}
	b.primaryColor = c
	b.updatedAt = now
	return nil
}

// UpdateSecondaryColor replaces the secondary accent color; empty
// clears it.
func (b *BrandSettings) UpdateSecondaryColor(c string, now time.Time) error {
	if err := validateColor(c); err != nil {
		return err
	}
	b.secondaryColor = c
	b.updatedAt = now
	return nil
}

func validateAppName(name string) error {
	if name == "" || len(name) > MaxAppNameLen {
		return ErrInvalidAppName
	}
	return nil
}

func validateTagline(tagline string) error {
	if len(tagline) > MaxTaglineLen {
		return ErrInvalidTagline
	}
	return nil
}

func validateColor(c string) error {
	if c == "" {
		return nil
	}
	if !hexColorRE.MatchString(c) {
		return ErrInvalidColor
	}
	return nil
}

func validateURL(u string) error {
	if u == "" {
		return nil
	}
	// strings.HasPrefix would be cheap but url.Parse catches subtler
	// malformations (whitespace, missing host, malformed escape
	// sequences) — let it do the heavy lifting then assert the
	// scheme whitelist.
	parsed, err := url.Parse(u)
	if err != nil {
		return ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}
	// Reject schemes that parse-successful but technically lack a
	// host (e.g., "http:") — defense-in-depth for the renderer.
	if parsed.Host == "" {
		return ErrInvalidURL
	}
	_ = strings.TrimSpace // reserved for future trimming policy
	return nil
}
