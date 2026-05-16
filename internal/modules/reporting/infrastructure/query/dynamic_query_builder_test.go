package query

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/xuri/excelize/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

// fieldNameID is the literal column-key string repeated across this test
// file's fixtures. Extracted as a const so the package-wide goconst threshold
// (min-occurrences: 25 in .github/golangci.yml) is not tripped — production
// dynamic_query_builder.go uses "name" 6 times legitimately as column-key,
// and the test file adds ~20 fixture references; pinning all test
// occurrences to this const keeps literal count well under the limit.
const fieldNameID = "name"

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
			mustHaveIDs:    []string{"id", fieldNameID, "category", "status", "size", "created_at", "updated_at", "author_name", "tags"},
			mustHaveEnumOn: "status",
		},
		{
			name:           "users",
			dataSource:     entities.DataSourceUsers,
			wantCount:      7,
			mustHaveIDs:    []string{"id", fieldNameID, "email", "role", "department", "created_at", "is_active"},
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
			mustHaveIDs:    []string{"id", fieldNameID, "group", "course", "faculty", "status", "enrolled_at"},
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

// ============================================================================
// B2 — buildWhereClause table-driven (15 operators + array branches + skips)
// ============================================================================

func TestBuildWhereClause_SingleFilterPerOperator(t *testing.T) {
	cfg := DataSourceConfig{
		TableName: "users u",
		ColumnMappings: map[string]string{
			fieldNameID: "u.name",
		},
	}

	cases := []struct {
		name       string
		operator   entities.FilterOperator
		value      any
		value2     any
		wantClause string
		wantArgs   []any
	}{
		{"equals", entities.FilterEquals, "Alice", nil, "u.name = $1", []any{"Alice"}},
		{"not_equals", entities.FilterNotEquals, "Alice", nil, "u.name != $1", []any{"Alice"}},
		{"contains", entities.FilterContains, "ali", nil, "u.name ILIKE $1", []any{"%ali%"}},
		{"not_contains", entities.FilterNotContains, "ali", nil, "u.name NOT ILIKE $1", []any{"%ali%"}},
		{"starts_with", entities.FilterStartsWith, "Al", nil, "u.name ILIKE $1", []any{"Al%"}},
		{"ends_with", entities.FilterEndsWith, "ce", nil, "u.name ILIKE $1", []any{"%ce"}},
		{"greater_than", entities.FilterGreaterThan, 10, nil, "u.name > $1", []any{10}},
		{"less_than", entities.FilterLessThan, 5, nil, "u.name < $1", []any{5}},
		{"greater_or_equal", entities.FilterGreaterOrEqual, 10, nil, "u.name >= $1", []any{10}},
		{"less_or_equal", entities.FilterLessOrEqual, 5, nil, "u.name <= $1", []any{5}},
		{"between", entities.FilterBetween, 1, 10, "u.name BETWEEN $1 AND $2", []any{1, 10}},
		{"is_null", entities.FilterIsNull, nil, nil, "u.name IS NULL", []any{}},
		{"is_not_null", entities.FilterIsNotNull, nil, nil, "u.name IS NOT NULL", []any{}},
	}

	b := &DynamicQueryBuilder{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filters := []entities.ReportFilterConfig{{
				Field:    entities.ReportField{Name: fieldNameID},
				Operator: tc.operator,
				Value:    tc.value,
				Value2:   tc.value2,
			}}
			clauses, args := b.buildWhereClause(filters, cfg)

			if len(clauses) != 1 {
				t.Fatalf("expected 1 clause, got %d: %v", len(clauses), clauses)
			}
			if clauses[0] != tc.wantClause {
				t.Errorf("clause = %q, want %q", clauses[0], tc.wantClause)
			}
			if len(args) != len(tc.wantArgs) {
				t.Fatalf("args len = %d, want %d (args=%v)", len(args), len(tc.wantArgs), args)
			}
			for i, want := range tc.wantArgs {
				if args[i] != want {
					t.Errorf("args[%d] = %v, want %v", i, args[i], want)
				}
			}
		})
	}
}

func TestBuildWhereClause_InAndNotInArrays(t *testing.T) {
	cfg := DataSourceConfig{
		ColumnMappings: map[string]string{"status": "u.status"},
	}
	b := &DynamicQueryBuilder{}

	cases := []struct {
		name       string
		operator   entities.FilterOperator
		value      any
		wantClause string
		wantArgs   []any
	}{
		{
			name:       "in_three_values",
			operator:   entities.FilterIn,
			value:      []any{"draft", "pending", "approved"},
			wantClause: "u.status IN ($1, $2, $3)",
			wantArgs:   []any{"draft", "pending", "approved"},
		},
		{
			name:       "not_in_two_values",
			operator:   entities.FilterNotIn,
			value:      []any{"rejected", "expelled"},
			wantClause: "u.status NOT IN ($1, $2)",
			wantArgs:   []any{"rejected", "expelled"},
		},
		{
			name:       "in_non_array_value_dropped",
			operator:   entities.FilterIn,
			value:      "not_an_array",
			wantClause: "", // skipped
			wantArgs:   []any{},
		},
		{
			name:       "in_empty_array_emits_degenerate_clause",
			operator:   entities.FilterIn,
			value:      []any{},
			wantClause: "u.status IN ()",
			wantArgs:   []any{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filters := []entities.ReportFilterConfig{{
				Field:    entities.ReportField{Name: "status"},
				Operator: tc.operator,
				Value:    tc.value,
			}}
			clauses, args := b.buildWhereClause(filters, cfg)

			if tc.wantClause == "" {
				if len(clauses) != 0 {
					t.Fatalf("expected no clauses, got %v", clauses)
				}
			} else {
				if len(clauses) != 1 {
					t.Fatalf("expected 1 clause, got %d: %v", len(clauses), clauses)
				}
				if clauses[0] != tc.wantClause {
					t.Errorf("clause = %q, want %q", clauses[0], tc.wantClause)
				}
			}
			if len(args) != len(tc.wantArgs) {
				t.Fatalf("args len = %d, want %d (%v)", len(args), len(tc.wantArgs), args)
			}
			for i, want := range tc.wantArgs {
				if args[i] != want {
					t.Errorf("args[%d] = %v, want %v", i, args[i], want)
				}
			}
		})
	}
}

func TestBuildWhereClause_UnknownFieldSkipped(t *testing.T) {
	cfg := DataSourceConfig{ColumnMappings: map[string]string{fieldNameID: "u.name"}}
	b := &DynamicQueryBuilder{}

	filters := []entities.ReportFilterConfig{
		{Field: entities.ReportField{Name: "unknown_field"}, Operator: entities.FilterEquals, Value: "x"},
		{Field: entities.ReportField{Name: fieldNameID}, Operator: entities.FilterEquals, Value: "Bob"},
	}
	clauses, args := b.buildWhereClause(filters, cfg)

	if len(clauses) != 1 {
		t.Fatalf("expected unknown field skipped, got clauses %v", clauses)
	}
	if clauses[0] != "u.name = $1" {
		t.Errorf("clause = %q, want %q", clauses[0], "u.name = $1")
	}
	if len(args) != 1 || args[0] != "Bob" {
		t.Errorf("args = %v, want [Bob]", args)
	}
}

func TestBuildWhereClause_UnknownOperatorSkipped(t *testing.T) {
	cfg := DataSourceConfig{ColumnMappings: map[string]string{fieldNameID: "u.name"}}
	b := &DynamicQueryBuilder{}

	filters := []entities.ReportFilterConfig{{
		Field:    entities.ReportField{Name: fieldNameID},
		Operator: entities.FilterOperator("nonsense_op"),
		Value:    "x",
	}}
	clauses, args := b.buildWhereClause(filters, cfg)

	if len(clauses) != 0 {
		t.Errorf("expected unknown operator skipped, got %v", clauses)
	}
	if len(args) != 0 {
		t.Errorf("expected no args, got %v", args)
	}
}

func TestBuildWhereClause_ArgIndexIncrementsAcrossFilters(t *testing.T) {
	cfg := DataSourceConfig{
		ColumnMappings: map[string]string{
			fieldNameID: "u.name",
			"status":    "u.status",
			"age":       "u.age",
		},
	}
	b := &DynamicQueryBuilder{}

	filters := []entities.ReportFilterConfig{
		{Field: entities.ReportField{Name: fieldNameID}, Operator: entities.FilterEquals, Value: "Alice"},
		{Field: entities.ReportField{Name: "age"}, Operator: entities.FilterBetween, Value: 18, Value2: 65},
		{Field: entities.ReportField{Name: "status"}, Operator: entities.FilterIn, Value: []any{"active", "pending"}},
	}
	clauses, args := b.buildWhereClause(filters, cfg)

	wantClauses := []string{
		"u.name = $1",
		"u.age BETWEEN $2 AND $3",
		"u.status IN ($4, $5)",
	}
	if len(clauses) != len(wantClauses) {
		t.Fatalf("clauses len = %d, want %d: %v", len(clauses), len(wantClauses), clauses)
	}
	for i, want := range wantClauses {
		if clauses[i] != want {
			t.Errorf("clauses[%d] = %q, want %q", i, clauses[i], want)
		}
	}

	wantArgs := []any{"Alice", 18, 65, "active", "pending"}
	if len(args) != len(wantArgs) {
		t.Fatalf("args len = %d, want %d: %v", len(args), len(wantArgs), args)
	}
	for i, want := range wantArgs {
		if args[i] != want {
			t.Errorf("args[%d] = %v, want %v", i, args[i], want)
		}
	}
}

// ============================================================================
// B3 — Execute via sqlmock (happy + error branches + composite paths)
// ============================================================================

// newMockBuilder bootstraps a DynamicQueryBuilder wired к sqlmock DB.
// Always defer db.Close() от caller.
func newMockBuilder(t *testing.T) (*DynamicQueryBuilder, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	return NewDynamicQueryBuilder(db), mock, db
}

func reportForDocuments(field, alias string, agg entities.AggregationType) *entities.CustomReport {
	return &entities.CustomReport{
		DataSource: entities.DataSourceDocuments,
		Fields: []entities.SelectedField{{
			Field:       entities.ReportField{Name: field, Label: "Lbl-" + field},
			Alias:       alias,
			Aggregation: agg,
		}},
	}
}

func anchor(q string) string {
	return "^" + regexp.QuoteMeta(q) + "$"
}

func TestExecute_UnsupportedDataSource(t *testing.T) {
	b, _, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	report := &entities.CustomReport{
		DataSource: entities.DataSourceType("nonexistent"),
		Fields:     []entities.SelectedField{{Field: entities.ReportField{Name: "id"}}},
	}
	_, err := b.Execute(context.Background(), report, 1, 10)
	if err == nil || !contains(err.Error(), "unsupported data source") {
		t.Fatalf("expected 'unsupported data source' error, got %v", err)
	}
}

func TestExecute_NoValidFields(t *testing.T) {
	b, _, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	report := &entities.CustomReport{
		DataSource: entities.DataSourceUsers,
		Fields: []entities.SelectedField{{
			Field: entities.ReportField{Name: "this_field_does_not_exist"},
		}},
	}
	_, err := b.Execute(context.Background(), report, 1, 10)
	if err == nil || !contains(err.Error(), "no valid fields selected") {
		t.Fatalf("expected 'no valid fields selected' error, got %v", err)
	}
}

func TestExecute_HappyPathSingleRow(t *testing.T) {
	b, mock, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(anchor("SELECT COUNT(*) FROM documents d LEFT JOIN users u ON d.author_id = u.id")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))
	mock.ExpectQuery(anchor("SELECT d.name AS name FROM documents d LEFT JOIN users u ON d.author_id = u.id ORDER BY 1 ASC LIMIT 10 OFFSET 0")).
		WillReturnRows(sqlmock.NewRows([]string{fieldNameID}).
			AddRow("Doc-A").
			AddRow("Doc-B").
			AddRow("Doc-C"))

	report := reportForDocuments(fieldNameID, "", entities.AggregationNone)
	result, err := b.Execute(context.Background(), report, 1, 10)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", result.TotalCount)
	}
	if result.Page != 1 || result.PageSize != 10 || result.TotalPages != 1 {
		t.Errorf("pagination = (%d/%d/%d), want (1/10/1)", result.Page, result.PageSize, result.TotalPages)
	}
	if len(result.Rows) != 3 {
		t.Fatalf("Rows len = %d, want 3", len(result.Rows))
	}
	if got := result.Rows[0][fieldNameID]; got != "Doc-A" {
		t.Errorf("Rows[0][name] = %v, want Doc-A", got)
	}
	if len(result.Columns) != 1 || result.Columns[0].Key != fieldNameID || result.Columns[0].Label != "Lbl-name" {
		t.Errorf("Columns = %+v, want [{name, Lbl-name}]", result.Columns)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations unmet: %v", err)
	}
}

func TestExecute_AggregationBranches(t *testing.T) {
	cases := []struct {
		name        string
		agg         entities.AggregationType
		wantColExpr string
	}{
		{"count", entities.AggregationCount, "COUNT(d.name)"},
		{"sum", entities.AggregationSum, "SUM(d.name)"},
		{"avg", entities.AggregationAvg, "AVG(d.name)"},
		{"min", entities.AggregationMin, "MIN(d.name)"},
		{"max", entities.AggregationMax, "MAX(d.name)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, mock, db := newMockBuilder(t)
			defer func() { _ = db.Close() }()

			mock.ExpectQuery(anchor("SELECT COUNT(*) FROM documents d LEFT JOIN users u ON d.author_id = u.id")).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
			mock.ExpectQuery(anchor("SELECT " + tc.wantColExpr + " AS name FROM documents d LEFT JOIN users u ON d.author_id = u.id ORDER BY 1 ASC LIMIT 5 OFFSET 0")).
				WillReturnRows(sqlmock.NewRows([]string{fieldNameID}))

			report := reportForDocuments(fieldNameID, "", tc.agg)
			result, err := b.Execute(context.Background(), report, 1, 5)
			if err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if result.TotalCount != 0 {
				t.Errorf("TotalCount = %d, want 0", result.TotalCount)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations unmet: %v", err)
			}
		})
	}
}

func TestExecute_AliasReplacesFieldName(t *testing.T) {
	b, mock, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(anchor("SELECT COUNT(*) FROM documents d LEFT JOIN users u ON d.author_id = u.id")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(anchor("SELECT d.name AS title FROM documents d LEFT JOIN users u ON d.author_id = u.id ORDER BY 1 ASC LIMIT 10 OFFSET 0")).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("X"))

	report := reportForDocuments(fieldNameID, "title", entities.AggregationNone)
	result, err := b.Execute(context.Background(), report, 1, 10)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Columns[0].Key != "title" {
		t.Errorf("Columns[0].Key = %q, want title", result.Columns[0].Key)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations unmet: %v", err)
	}
}

func TestExecute_WithGroupBySortByDescPaginationAndFilters(t *testing.T) {
	b, mock, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	expectedCount := "SELECT COUNT(*) FROM documents d LEFT JOIN users u ON d.author_id = u.id WHERE d.status = $1"
	expectedMain := "SELECT d.status AS status FROM documents d LEFT JOIN users u ON d.author_id = u.id WHERE d.status = $1 GROUP BY d.status ORDER BY d.status DESC LIMIT 5 OFFSET 5"

	mock.ExpectQuery(anchor(expectedCount)).WithArgs("approved").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(12)))
	mock.ExpectQuery(anchor(expectedMain)).WithArgs("approved").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).
			AddRow("approved").
			AddRow("approved"))

	report := &entities.CustomReport{
		DataSource: entities.DataSourceDocuments,
		Fields: []entities.SelectedField{{
			Field: entities.ReportField{Name: "status", Label: "Status"},
		}},
		Filters: []entities.ReportFilterConfig{{
			Field:    entities.ReportField{Name: "status"},
			Operator: entities.FilterEquals,
			Value:    "approved",
		}},
		Groupings: []entities.ReportGrouping{{
			Field: entities.ReportField{Name: "status"},
		}, {
			Field: entities.ReportField{Name: "unknown_group_skipped"},
		}},
		Sortings: []entities.ReportSorting{{
			Field: entities.ReportField{Name: "status"},
			Order: entities.SortOrderDesc,
		}, {
			Field: entities.ReportField{Name: "unknown_sort_skipped"},
		}},
	}

	result, err := b.Execute(context.Background(), report, 2, 5)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.TotalCount != 12 {
		t.Errorf("TotalCount = %d, want 12", result.TotalCount)
	}
	if result.TotalPages != 3 { // ceil(12/5)
		t.Errorf("TotalPages = %d, want 3", result.TotalPages)
	}
	if result.Page != 2 || result.PageSize != 5 {
		t.Errorf("pagination = (%d/%d), want (2/5)", result.Page, result.PageSize)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations unmet: %v", err)
	}
}

func TestExecute_DefaultOrderByWhenNoSortings(t *testing.T) {
	b, mock, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(anchor("SELECT COUNT(*) FROM users u")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(anchor("SELECT u.email AS email FROM users u ORDER BY 1 ASC LIMIT 10 OFFSET 0")).
		WillReturnRows(sqlmock.NewRows([]string{"email"}))

	report := &entities.CustomReport{
		DataSource: entities.DataSourceUsers,
		Fields:     []entities.SelectedField{{Field: entities.ReportField{Name: "email"}}},
	}
	_, err := b.Execute(context.Background(), report, 1, 10)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations unmet: %v", err)
	}
}

func TestExecute_CountQueryError(t *testing.T) {
	b, mock, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(anchor("SELECT COUNT(*) FROM users u")).
		WillReturnError(errors.New("count boom"))

	report := &entities.CustomReport{
		DataSource: entities.DataSourceUsers,
		Fields:     []entities.SelectedField{{Field: entities.ReportField{Name: "email"}}},
	}
	_, err := b.Execute(context.Background(), report, 1, 10)
	if err == nil || !contains(err.Error(), "failed to count records") {
		t.Fatalf("expected 'failed to count records' error, got %v", err)
	}
}

func TestExecute_MainQueryError(t *testing.T) {
	b, mock, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(anchor("SELECT COUNT(*) FROM users u")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))
	mock.ExpectQuery(anchor("SELECT u.email AS email FROM users u ORDER BY 1 ASC LIMIT 10 OFFSET 0")).
		WillReturnError(errors.New("select boom"))

	report := &entities.CustomReport{
		DataSource: entities.DataSourceUsers,
		Fields:     []entities.SelectedField{{Field: entities.ReportField{Name: "email"}}},
	}
	_, err := b.Execute(context.Background(), report, 1, 10)
	if err == nil || !contains(err.Error(), "failed to execute query") {
		t.Fatalf("expected 'failed to execute query' error, got %v", err)
	}
}

// Note: `failed to scan row` error path не покрывается стандартным sqlmock.
// rows.Scan() error через RowError приземляется в rows.Err() после Next()
// returns false — Execute() не checks rows.Err() после loop, поэтому
// error path silently drops. Coverage delta для этой строки = 1 statement.
// Carry-forward: либо refactor Execute() добавить rows.Err() check (production
// behavior change — out of scope для backfill), либо использовать custom
// driver. Принято accept ~2 LoC uncovered.

func TestExecute_ByteArrayValuesConvertedToString(t *testing.T) {
	b, mock, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(anchor("SELECT COUNT(*) FROM users u")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(anchor("SELECT u.email AS email FROM users u ORDER BY 1 ASC LIMIT 10 OFFSET 0")).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow([]byte("byte-stream@x.io")))

	report := &entities.CustomReport{
		DataSource: entities.DataSourceUsers,
		Fields:     []entities.SelectedField{{Field: entities.ReportField{Name: "email"}}},
	}
	result, err := b.Execute(context.Background(), report, 1, 10)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := result.Rows[0]["email"]; got != "byte-stream@x.io" {
		t.Errorf("Rows[0][email] = %v (%T), want string %q", got, got, "byte-stream@x.io")
	}
}

func TestExecute_NilValuePassesThroughMap(t *testing.T) {
	b, mock, db := newMockBuilder(t)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(anchor("SELECT COUNT(*) FROM users u")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(anchor("SELECT u.email AS email FROM users u ORDER BY 1 ASC LIMIT 10 OFFSET 0")).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow(nil))

	report := &entities.CustomReport{
		DataSource: entities.DataSourceUsers,
		Fields:     []entities.SelectedField{{Field: entities.ReportField{Name: "email"}}},
	}
	result, err := b.Execute(context.Background(), report, 1, 10)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := result.Rows[0]["email"]; got != nil {
		t.Errorf("Rows[0][email] = %v (%T), want nil", got, got)
	}
}

// contains is a thin wrapper for error-message substring assertions.
func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// ============================================================================
// B4 — Export + exportCSV + exportXLSX + exportPDF
// ============================================================================

func sampleResult() *entities.ReportExecutionResult {
	return &entities.ReportExecutionResult{
		Columns: []entities.ReportColumn{
			{Key: fieldNameID, Label: "Name"},
			{Key: "count", Label: "Count"},
		},
		Rows: []map[string]any{
			{fieldNameID: "Alpha", "count": int64(10)},
			{fieldNameID: "Beta", "count": int64(20)},
		},
		TotalCount: 2, Page: 1, PageSize: 10, TotalPages: 1,
	}
}

func TestExport_UnsupportedFormat(t *testing.T) {
	b := &DynamicQueryBuilder{}
	_, _, err := b.Export(sampleResult(), entities.ExportOptions{Format: "rtf"}, "rep")
	if err == nil || !strings.Contains(err.Error(), "unsupported export format") {
		t.Fatalf("expected 'unsupported export format', got %v", err)
	}
}

func TestExportCSV_WithHeaders(t *testing.T) {
	b := &DynamicQueryBuilder{}
	data, filename, err := b.Export(
		sampleResult(),
		entities.ExportOptions{Format: entities.ExportFormatCSV, IncludeHeaders: true},
		"My Report",
	)
	if err != nil {
		t.Fatalf("Export csv: %v", err)
	}
	if !strings.HasPrefix(filename, "My_Report_") || !strings.HasSuffix(filename, ".csv") {
		t.Errorf("filename = %q, want prefix 'My_Report_' suffix '.csv'", filename)
	}
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll: %v", err)
	}
	want := [][]string{
		{"Name", "Count"},
		{"Alpha", "10"},
		{"Beta", "20"},
	}
	if len(records) != len(want) {
		t.Fatalf("records len = %d, want %d (records=%v)", len(records), len(want), records)
	}
	for i, row := range want {
		if len(records[i]) != len(row) {
			t.Fatalf("row %d cols = %d, want %d", i, len(records[i]), len(row))
		}
		for j, cell := range row {
			if records[i][j] != cell {
				t.Errorf("record[%d][%d] = %q, want %q", i, j, records[i][j], cell)
			}
		}
	}
}

func TestExportCSV_WithoutHeaders(t *testing.T) {
	b := &DynamicQueryBuilder{}
	data, _, err := b.Export(
		sampleResult(),
		entities.ExportOptions{Format: entities.ExportFormatCSV, IncludeHeaders: false},
		"r",
	)
	if err != nil {
		t.Fatalf("Export csv: %v", err)
	}
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll: %v", err)
	}
	// IncludeHeaders=false → exactly len(Rows) records, no header line.
	if len(records) != 2 {
		t.Fatalf("records len = %d, want 2 (no-header expectation): %v", len(records), records)
	}
	want := [][]string{
		{"Alpha", "10"},
		{"Beta", "20"},
	}
	for i, row := range want {
		for j, cell := range row {
			if records[i][j] != cell {
				t.Errorf("record[%d][%d] = %q, want %q", i, j, records[i][j], cell)
			}
		}
	}
}

func TestExportCSV_FormatsValueTypes(t *testing.T) {
	result := &entities.ReportExecutionResult{
		Columns: []entities.ReportColumn{
			{Key: "ts", Label: "TS"},
			{Key: "flag", Label: "Flag"},
			{Key: "missing", Label: "Missing"},
		},
		Rows: []map[string]any{{
			"ts":      time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC),
			"flag":    true,
			"missing": nil,
		}},
	}
	b := &DynamicQueryBuilder{}
	data, _, err := b.Export(result, entities.ExportOptions{Format: entities.ExportFormatCSV, IncludeHeaders: false}, "r")
	if err != nil {
		t.Fatalf("Export csv: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "2026-05-16 12:00:00") {
		t.Errorf("CSV missing time formatting:\n%s", got)
	}
	if !strings.Contains(got, "Yes") {
		t.Errorf("CSV missing bool true → 'Yes':\n%s", got)
	}
}

func TestExportXLSX_WithHeaders(t *testing.T) {
	b := &DynamicQueryBuilder{}
	data, filename, err := b.Export(
		sampleResult(),
		entities.ExportOptions{Format: entities.ExportFormatXLSX, IncludeHeaders: true},
		"Stats",
	)
	if err != nil {
		t.Fatalf("Export xlsx: %v", err)
	}
	if !strings.HasPrefix(filename, "Stats_") || !strings.HasSuffix(filename, ".xlsx") {
		t.Errorf("filename = %q", filename)
	}
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' || data[2] != 0x03 || data[3] != 0x04 {
		t.Fatalf("XLSX magic bytes mismatch: got %x", data[:4])
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("excelize.OpenReader: %v", err)
	}
	defer func() { _ = f.Close() }()

	wantCells := map[string]string{
		"A1": "Name",
		"B1": "Count",
		"A2": "Alpha",
		"B2": "10",
		"A3": "Beta",
		"B3": "20",
	}
	for axis, want := range wantCells {
		got, err := f.GetCellValue("Report", axis)
		if err != nil {
			t.Errorf("GetCellValue(Report,%s): %v", axis, err)
			continue
		}
		if got != want {
			t.Errorf("Report!%s = %q, want %q", axis, got, want)
		}
	}
}

func TestExportXLSX_WithoutHeaders(t *testing.T) {
	b := &DynamicQueryBuilder{}
	data, _, err := b.Export(
		sampleResult(),
		entities.ExportOptions{Format: entities.ExportFormatXLSX, IncludeHeaders: false},
		"r",
	)
	if err != nil {
		t.Fatalf("Export xlsx no-headers: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("excelize.OpenReader: %v", err)
	}
	defer func() { _ = f.Close() }()

	// IncludeHeaders=false → data starts at row 1, no header row.
	a1, _ := f.GetCellValue("Report", "A1")
	if a1 != "Alpha" {
		t.Errorf("Report!A1 = %q, want %q (no-header expectation)", a1, "Alpha")
	}
	b1, _ := f.GetCellValue("Report", "B1")
	if b1 != "10" {
		t.Errorf("Report!B1 = %q, want %q", b1, "10")
	}
}

func TestExportPDF_HappyPath(t *testing.T) {
	b := &DynamicQueryBuilder{}
	data, filename, err := b.Export(
		sampleResult(),
		entities.ExportOptions{Format: entities.ExportFormatPDF, IncludeHeaders: true},
		"PDF Report",
	)
	if err != nil {
		t.Fatalf("Export pdf: %v", err)
	}
	if !strings.HasPrefix(filename, "PDF_Report_") || !strings.HasSuffix(filename, ".pdf") {
		t.Errorf("filename = %q", filename)
	}
	if len(data) < 5 || !strings.HasPrefix(string(data), "%PDF-") {
		t.Errorf("PDF magic mismatch: got %q", string(data[:min(len(data), 8)]))
	}
}

func TestExportPDF_LandscapeOrientation(t *testing.T) {
	b := &DynamicQueryBuilder{}
	data, _, err := b.Export(
		sampleResult(),
		entities.ExportOptions{
			Format:         entities.ExportFormatPDF,
			IncludeHeaders: true,
			Orientation:    "landscape",
		},
		"L",
	)
	if err != nil {
		t.Fatalf("Export pdf landscape: %v", err)
	}
	if len(data) < 5 || !strings.HasPrefix(string(data), "%PDF-") {
		t.Errorf("PDF magic mismatch: got %q", string(data[:min(len(data), 8)]))
	}
}

func TestExportPDF_CustomPageSize(t *testing.T) {
	b := &DynamicQueryBuilder{}
	data, _, err := b.Export(
		sampleResult(),
		entities.ExportOptions{
			Format:         entities.ExportFormatPDF,
			IncludeHeaders: true,
			PageSize:       "Letter",
		},
		"L",
	)
	if err != nil {
		t.Fatalf("Export pdf letter: %v", err)
	}
	if len(data) < 5 || !strings.HasPrefix(string(data), "%PDF-") {
		t.Errorf("PDF magic mismatch: got %q", string(data[:min(len(data), 8)]))
	}
}

func TestExportPDF_WithoutHeaders(t *testing.T) {
	b := &DynamicQueryBuilder{}
	data, _, err := b.Export(
		sampleResult(),
		entities.ExportOptions{
			Format:         entities.ExportFormatPDF,
			IncludeHeaders: false,
		},
		"r",
	)
	if err != nil {
		t.Fatalf("Export pdf no-headers: %v", err)
	}
	if len(data) < 5 || !strings.HasPrefix(string(data), "%PDF-") {
		t.Errorf("PDF magic mismatch: got %q", string(data[:min(len(data), 8)]))
	}
}

func TestExportPDF_NoColumnsError(t *testing.T) {
	b := &DynamicQueryBuilder{}
	emptyCols := &entities.ReportExecutionResult{
		Columns: []entities.ReportColumn{},
		Rows:    []map[string]any{},
	}
	_, _, err := b.Export(emptyCols, entities.ExportOptions{Format: entities.ExportFormatPDF}, "r")
	if err == nil || !strings.Contains(err.Error(), "no columns in report") {
		t.Fatalf("expected 'no columns in report' error, got %v", err)
	}
}

func TestExportPDF_PageBreakWithManyRows(t *testing.T) {
	// Many rows to trigger pdf.GetY() > 270 page-break branch.
	rows := make([]map[string]any, 0, 50)
	for i := range 50 {
		rows = append(rows, map[string]any{
			fieldNameID: "Row",
			"count":     int64(i),
		})
	}
	result := &entities.ReportExecutionResult{
		Columns: []entities.ReportColumn{
			{Key: fieldNameID, Label: "Name"},
			{Key: "count", Label: "Count"},
		},
		Rows: rows,
	}
	b := &DynamicQueryBuilder{}
	data, _, err := b.Export(result, entities.ExportOptions{
		Format:         entities.ExportFormatPDF,
		IncludeHeaders: true,
	}, "many")
	if err != nil {
		t.Fatalf("Export pdf many rows: %v", err)
	}
	if len(data) < 5 || !strings.HasPrefix(string(data), "%PDF-") {
		t.Errorf("PDF magic mismatch: got %q", string(data[:min(len(data), 8)]))
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
