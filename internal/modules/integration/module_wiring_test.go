package integration

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/infrastructure/odata"
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
		DebtCatalog:     "Catalog_АкадемическиеЗадолженности",
	}

	got := buildODataConfig(cfg)

	if got.EmployeesCatalog != cfg.EmployeeCatalog {
		t.Errorf("EmployeesCatalog = %q, want %q — empty means the client hits the OData root and syncs nothing",
			got.EmployeesCatalog, cfg.EmployeeCatalog)
	}
	if got.StudentsCatalog != cfg.StudentCatalog {
		t.Errorf("StudentsCatalog = %q, want %q", got.StudentsCatalog, cfg.StudentCatalog)
	}
	if got.StudentDebtsCatalog != cfg.DebtCatalog {
		t.Errorf("StudentDebtsCatalog = %q, want %q — empty means the 1С debt fetch hits the OData root and imports nothing",
			got.StudentDebtsCatalog, cfg.DebtCatalog)
	}

	// Sanity: the previously-mapped fields still carry through.
	if got.BaseURL != cfg.BaseURL || got.Username != cfg.Username ||
		got.MaxRetries != cfg.MaxRetries || got.RetryDelay != cfg.RetryDelay {
		t.Errorf("base fields not mapped correctly: %+v", got)
	}
}

// TestModule_ODataClient guards the accessor the DI layer uses to share the 1C
// client with cross-module adapters (student_debts 1С debt import): a disabled
// module exposes nil; a module carrying a client returns it.
func TestModule_ODataClient(t *testing.T) {
	disabled := &Module{}
	if disabled.ODataClient() != nil {
		t.Error("a disabled module must expose a nil OData client")
	}

	client := odata.NewClient(odata.DefaultConfig())
	m := &Module{odataClient: client}
	if m.ODataClient() != client {
		t.Error("ODataClient must return the stored client")
	}
}
