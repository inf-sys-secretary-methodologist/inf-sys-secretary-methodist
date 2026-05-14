package entities

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedNow = time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)

func TestNewBrandSettings_HappyPath(t *testing.T) {
	bs, err := NewBrandSettings(
		"Acme University",
		"Building the future, one student at a time",
		"https://example.com/logo.png",
		"https://example.com/favicon.ico",
		"#FF5733",
		"#0066CC",
		fixedNow,
	)
	require.NoError(t, err)
	require.NotNil(t, bs)
	assert.Equal(t, "Acme University", bs.AppName())
	assert.Equal(t, "Building the future, one student at a time", bs.Tagline())
	assert.Equal(t, "https://example.com/logo.png", bs.LogoURL())
	assert.Equal(t, "https://example.com/favicon.ico", bs.FaviconURL())
	assert.Equal(t, "#FF5733", bs.PrimaryColor())
	assert.Equal(t, "#0066CC", bs.SecondaryColor())
	assert.Equal(t, fixedNow, bs.UpdatedAt())
}

func TestNewBrandSettings_OptionalFieldsEmpty(t *testing.T) {
	bs, err := NewBrandSettings("Acme", "", "", "", "", "", fixedNow)
	require.NoError(t, err)
	require.NotNil(t, bs)
	assert.Equal(t, "Acme", bs.AppName())
	assert.Empty(t, bs.LogoURL())
	assert.Empty(t, bs.PrimaryColor())
}

func TestValidateAppName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty", "", ErrInvalidAppName},
		{"single char", "A", nil},
		{"at max length", strings.Repeat("x", MaxAppNameLen), nil},
		{"over max", strings.Repeat("x", MaxAppNameLen+1), ErrInvalidAppName},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewBrandSettings(tc.input, "", "", "", "", "", fixedNow)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected errors.Is(%v, %v)", err, tc.wantErr)
			}
		})
	}
}

func TestValidateTagline(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty allowed", "", nil},
		{"normal", "Welcome to our system", nil},
		{"at max length", strings.Repeat("y", MaxTaglineLen), nil},
		{"over max", strings.Repeat("y", MaxTaglineLen+1), ErrInvalidTagline},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewBrandSettings("Acme", tc.input, "", "", "", "", fixedNow)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.True(t, errors.Is(err, tc.wantErr))
			}
		})
	}
}

func TestValidateColor(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty allowed", "", nil},
		{"6-digit lowercase", "#abcdef", nil},
		{"6-digit uppercase", "#ABCDEF", nil},
		{"6-digit mixed", "#aB12cD", nil},
		{"3-digit", "#f0f", nil},
		{"missing hash", "abcdef", ErrInvalidColor},
		{"too short", "#ab", ErrInvalidColor},
		{"too long", "#abcdefa", ErrInvalidColor},
		{"non-hex chars", "#abcxyz", ErrInvalidColor},
		{"trailing space", "#abcdef ", ErrInvalidColor},
		{"named color", "red", ErrInvalidColor},
		{"rgba", "rgba(255,0,0,1)", ErrInvalidColor},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewBrandSettings("Acme", "", "", "", tc.input, "", fixedNow)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.True(t, errors.Is(err, tc.wantErr),
					"input=%q expected %v got %v", tc.input, tc.wantErr, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty allowed", "", nil},
		{"https", "https://example.com/logo.png", nil},
		{"http", "http://example.com/logo.png", nil},
		{"http with port", "http://localhost:9000/logo.png", nil},
		{"javascript scheme", "javascript:alert(1)", ErrInvalidURL},
		{"data scheme", "data:image/png;base64,iVBOR...", ErrInvalidURL},
		{"file scheme", "file:///etc/passwd", ErrInvalidURL},
		{"ftp scheme", "ftp://example.com/logo", ErrInvalidURL},
		{"no scheme", "example.com/logo.png", ErrInvalidURL},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewBrandSettings("Acme", "", tc.input, "", "", "", fixedNow)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.True(t, errors.Is(err, tc.wantErr),
					"input=%q expected %v got %v", tc.input, tc.wantErr, err)
			}
		})
	}
}

func TestUpdateAppName(t *testing.T) {
	bs, _ := NewBrandSettings("Acme", "", "", "", "", "", fixedNow)
	later := fixedNow.Add(time.Hour)

	err := bs.UpdateAppName("New Name", later)
	require.NoError(t, err)
	assert.Equal(t, "New Name", bs.AppName())
	assert.Equal(t, later, bs.UpdatedAt(), "successful update touches updatedAt")

	err = bs.UpdateAppName("", later.Add(time.Hour))
	assert.True(t, errors.Is(err, ErrInvalidAppName))
	assert.Equal(t, "New Name", bs.AppName(), "rejected update leaves field unchanged")
	assert.Equal(t, later, bs.UpdatedAt(), "rejected update leaves updatedAt unchanged")
}

func TestUpdateTagline(t *testing.T) {
	bs, _ := NewBrandSettings("Acme", "old", "", "", "", "", fixedNow)
	later := fixedNow.Add(time.Hour)

	require.NoError(t, bs.UpdateTagline("new tagline", later))
	assert.Equal(t, "new tagline", bs.Tagline())
	assert.Equal(t, later, bs.UpdatedAt())

	err := bs.UpdateTagline(strings.Repeat("y", MaxTaglineLen+1), later.Add(time.Hour))
	assert.True(t, errors.Is(err, ErrInvalidTagline))
	assert.Equal(t, "new tagline", bs.Tagline())
}

func TestUpdateLogoURL(t *testing.T) {
	bs, _ := NewBrandSettings("Acme", "", "https://old.example/logo", "", "", "", fixedNow)
	later := fixedNow.Add(time.Hour)

	require.NoError(t, bs.UpdateLogoURL("https://new.example/logo.png", later))
	assert.Equal(t, "https://new.example/logo.png", bs.LogoURL())

	err := bs.UpdateLogoURL("javascript:alert(1)", later.Add(time.Hour))
	assert.True(t, errors.Is(err, ErrInvalidURL))
	assert.Equal(t, "https://new.example/logo.png", bs.LogoURL())
}

func TestUpdateFaviconURL(t *testing.T) {
	bs, _ := NewBrandSettings("Acme", "", "", "", "", "", fixedNow)

	require.NoError(t, bs.UpdateFaviconURL("https://example.com/favicon.ico", fixedNow))
	assert.Equal(t, "https://example.com/favicon.ico", bs.FaviconURL())

	err := bs.UpdateFaviconURL("file:///etc/passwd", fixedNow)
	assert.True(t, errors.Is(err, ErrInvalidURL))
}

func TestUpdatePrimaryColor(t *testing.T) {
	bs, _ := NewBrandSettings("Acme", "", "", "", "", "", fixedNow)

	require.NoError(t, bs.UpdatePrimaryColor("#abc", fixedNow))
	assert.Equal(t, "#abc", bs.PrimaryColor())

	err := bs.UpdatePrimaryColor("not-a-color", fixedNow)
	assert.True(t, errors.Is(err, ErrInvalidColor))
}

func TestUpdateSecondaryColor(t *testing.T) {
	bs, _ := NewBrandSettings("Acme", "", "", "", "", "", fixedNow)

	require.NoError(t, bs.UpdateSecondaryColor("#FFAA00", fixedNow))
	assert.Equal(t, "#FFAA00", bs.SecondaryColor())

	err := bs.UpdateSecondaryColor("rgba(0,0,0,0)", fixedNow)
	assert.True(t, errors.Is(err, ErrInvalidColor))
}

func TestRehydrateBrandSettings_BypassesValidation(t *testing.T) {
	bs := RehydrateBrandSettings("loaded", "", "", "", "", "", fixedNow)
	assert.Equal(t, "loaded", bs.AppName())
	assert.Equal(t, fixedNow, bs.UpdatedAt())
}
