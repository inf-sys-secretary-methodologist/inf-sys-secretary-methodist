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
// only on successful validation. RED stub returns nil без mutation —
// GREEN restores validation + write.
func (b *BrandSettings) UpdateAppName(_ string, _ time.Time) error {
	return nil
}

// UpdateTagline — stub в RED commit.
func (b *BrandSettings) UpdateTagline(_ string, _ time.Time) error {
	return nil
}

// UpdateLogoURL — stub в RED commit.
func (b *BrandSettings) UpdateLogoURL(_ string, _ time.Time) error {
	return nil
}

// UpdateFaviconURL — stub в RED commit.
func (b *BrandSettings) UpdateFaviconURL(_ string, _ time.Time) error {
	return nil
}

// UpdatePrimaryColor — stub в RED commit.
func (b *BrandSettings) UpdatePrimaryColor(_ string, _ time.Time) error {
	return nil
}

// UpdateSecondaryColor — stub в RED commit.
func (b *BrandSettings) UpdateSecondaryColor(_ string, _ time.Time) error {
	return nil
}

// validateAppName — RED stub does no validation so NewBrandSettings
// happy path passes; invariant tests on empty / too-long names fail
// until GREEN ships the real predicate.
func validateAppName(_ string) error {
	return nil
}

func validateTagline(_ string) error {
	return nil
}

func validateColor(_ string) error {
	return nil
}

func validateURL(_ string) error {
	return nil
}

// References reserved for the GREEN impl — silence "unused" linter
// until the real predicates land in the next commit.
var (
	_ = hexColorRE
	_ = strings.TrimSpace
	_ = url.Parse
)
