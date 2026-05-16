package query

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

// ============================================================================
// B1 — Pure helpers + GetAvailableFields
// ============================================================================

func TestFormatValue(t *testing.T) {
	fixedTime := time.Date(2026, 5, 16, 10, 30, 45, 0, time.UTC)

	cases := []struct {
		name string
		in   any
		want string
	}{
		{"nil returns empty", nil, ""},
		{"time formatted yyyy-mm-dd HH:MM:SS", fixedTime, "2026-05-16 10:30:45"},
		{"bool true returns Yes", true, "Yes"},
		{"bool false returns No", false, "No"},
		{"int default Sprint", 42, "42"},
		{"string default passthrough", "hello", "hello"},
		{"int64 default Sprint", int64(9223372036854775807), "9223372036854775807"},
		{"float default Sprint", 3.14, "3.14"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatValue(tc.in)
			if got != tc.want {
				t.Errorf("formatValue(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"already_safe", "report_2026", "report_2026"},
		{"slash", "a/b", "a_b"},
		{"backslash", "a\\b", "a_b"},
		{"colon", "a:b", "a_b"},
		{"asterisk", "a*b", "a_b"},
		{"question", "a?b", "a_b"},
		{"quote", "a\"b", "a_b"},
		{"angle_less", "a<b", "a_b"},
		{"angle_greater", "a>b", "a_b"},
		{"pipe", "a|b", "a_b"},
		{"space", "a b", "a_b"},
		{"all_at_once", "a/b\\c:d*e?f\"g<h>i|j k", "a_b_c_d_e_f_g_h_i_j_k"},
		{"cyrillic_preserved", "Отчёт_2026", "Отчёт_2026"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeFilename(tc.in)
			if got != tc.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		maxLen int
		want   string
	}{
		{"shorter_than_max", "hi", 10, "hi"},
		{"exactly_max", "hello", 5, "hello"},
		{"longer_truncated_with_ellipsis", "hello world", 8, "hello..."},
		{"long_only_3_chars_left_after_ellipsis", "abcdefghij", 6, "abc..."},
		{"single_char_over", "abcdef", 5, "ab..."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := truncateString(tc.in, tc.maxLen)
			if got != tc.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tc.in, tc.maxLen, got, tc.want)
			}
		})
	}
}

func TestGetAvailableFields(t *testing.T) {
	b := NewDynamicQueryBuilder(nil) // db not used by GetAvailableFields

	cases := []struct {
		name           string
		dataSource     entities.DataSourceType
		wantCount      int
		mustHaveIDs    []string
		mustHaveEnumOn string
	}{
		{
			name:           "documents",
			dataSource:     entities.DataSourceDocuments,
			wantCount:      9,
			mustHaveIDs:    []string{"id", "name", "category", "status", "size", "created_at", "updated_at", "author_name", "tags"},
			mustHaveEnumOn: "status",
		},
		{
			name:           "users",
			dataSource:     entities.DataSourceUsers,
			wantCount:      7,
			mustHaveIDs:    []string{"id", "name", "email", "role", "department", "created_at", "is_active"},
			mustHaveEnumOn: "role",
		},
		{
			name:           "events",
			dataSource:     entities.DataSourceEvents,
			wantCount:      7,
			mustHaveIDs:    []string{"id", "title", "type", "start_time", "end_time", "location", "organizer"},
			mustHaveEnumOn: "type",
		},
		{
			name:           "tasks",
			dataSource:     entities.DataSourceTasks,
			wantCount:      7,
			mustHaveIDs:    []string{"id", "title", "status", "priority", "due_date", "assignee", "created_at"},
			mustHaveEnumOn: "status",
		},
		{
			name:           "students",
			dataSource:     entities.DataSourceStudents,
			wantCount:      7,
			mustHaveIDs:    []string{"id", "name", "group", "course", "faculty", "status", "enrolled_at"},
			mustHaveEnumOn: "status",
		},
		{
			name:       "unknown_returns_empty",
			dataSource: entities.DataSourceType("unknown"),
			wantCount:  0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fields := b.GetAvailableFields(tc.dataSource)
			if len(fields) != tc.wantCount {
				t.Fatalf("GetAvailableFields(%s) returned %d fields, want %d", tc.dataSource, len(fields), tc.wantCount)
			}
			seen := map[string]entities.ReportField{}
			for _, f := range fields {
				seen[f.ID] = f
			}
			for _, id := range tc.mustHaveIDs {
				f, ok := seen[id]
				if !ok {
					t.Errorf("GetAvailableFields(%s) missing field ID %q", tc.dataSource, id)
					continue
				}
				if f.Source != tc.dataSource {
					t.Errorf("field %q has Source=%q, want %q", id, f.Source, tc.dataSource)
				}
			}
			if tc.mustHaveEnumOn != "" {
				f, ok := seen[tc.mustHaveEnumOn]
				if !ok {
					t.Fatalf("field %q expected on data source %q", tc.mustHaveEnumOn, tc.dataSource)
				}
				if f.Type != entities.FieldTypeEnum {
					t.Errorf("field %q.Type = %q, want enum", tc.mustHaveEnumOn, f.Type)
				}
				if len(f.EnumValues) == 0 {
					t.Errorf("field %q.EnumValues empty, want at least one", tc.mustHaveEnumOn)
				}
			}
		})
	}
}

func TestNewDynamicQueryBuilder_PopulatesAllSources(t *testing.T) {
	b := NewDynamicQueryBuilder(nil)

	wantSources := []entities.DataSourceType{
		entities.DataSourceDocuments,
		entities.DataSourceUsers,
		entities.DataSourceEvents,
		entities.DataSourceTasks,
		entities.DataSourceStudents,
	}

	for _, src := range wantSources {
		cfg, ok := b.sourceConfigs[src]
		if !ok {
			t.Errorf("sourceConfigs missing entry for %s", src)
			continue
		}
		if cfg.TableName == "" {
			t.Errorf("sourceConfigs[%s].TableName empty", src)
		}
		if len(cfg.ColumnMappings) == 0 {
			t.Errorf("sourceConfigs[%s].ColumnMappings empty", src)
		}
	}
}
