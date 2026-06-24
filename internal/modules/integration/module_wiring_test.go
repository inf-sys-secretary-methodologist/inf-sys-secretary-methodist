package integration

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
)

// TestBuildODataConfig_MapsCatalogNames guards the wiring bug where the 1C
// catalog names were never passed to the OData client, so every sync fetched
// the service root and imported zero rows. Surfaced by an end-to-end sync
// against a local mock-1C; the existing unit tests missed it because they
// inject a fake OData client and bypass this mapping.
func TestBuildODataConfig_MapsCatalogNames(t *testing.T) {
	cfg := &config.IntegrationConfig{
		BaseURL:         "http://1c/odata/standard.odata",
		Username:        "u",
		Password:        "p",
		Timeout:         5 * time.Second,
		MaxRetries:      2,
		RetryDelay:      time.Second,
		EmployeeCatalog: "Catalog_Сотрудники",
		StudentCatalog:  "Catalog_Студенты",
	}

	got := buildODataConfig(cfg)

	if got.EmployeesCatalog != cfg.EmployeeCatalog {
		t.Errorf("EmployeesCatalog = %q, want %q — empty means the client hits the OData root and syncs nothing",
			got.EmployeesCatalog, cfg.EmployeeCatalog)
	}
	if got.StudentsCatalog != cfg.StudentCatalog {
		t.Errorf("StudentsCatalog = %q, want %q", got.StudentsCatalog, cfg.StudentCatalog)
	}

	// Sanity: the previously-mapped fields still carry through.
	if got.BaseURL != cfg.BaseURL || got.Username != cfg.Username ||
		got.MaxRetries != cfg.MaxRetries || got.RetryDelay != cfg.RetryDelay {
		t.Errorf("base fields not mapped correctly: %+v", got)
	}
}
