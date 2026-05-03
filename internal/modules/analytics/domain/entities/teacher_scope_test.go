package entities_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

func TestNewTeacherScope_StoresTeacherIDAndDeduplicatesGroups(t *testing.T) {
	s := entities.NewTeacherScope(42, []string{"ИС-21", "ИС-22", "ИС-21", ""})

	require.NotNil(t, s)
	assert.Equal(t, int64(42), s.TeacherID())
	assert.True(t, s.AllowsGroup("ИС-21"))
	assert.True(t, s.AllowsGroup("ИС-22"))
	assert.False(t, s.AllowsGroup(""), "empty group must always be denied")
}

func TestTeacherScope_AllowsGroup(t *testing.T) {
	tests := []struct {
		name      string
		whitelist []string
		query     string
		want      bool
	}{
		{name: "match", whitelist: []string{"ИС-21", "ИС-22"}, query: "ИС-21", want: true},
		{name: "no match", whitelist: []string{"ИС-21"}, query: "ИС-99", want: false},
		{name: "nil whitelist denies any non-empty query", whitelist: nil, query: "ИС-21", want: false},
		{name: "empty query is denied", whitelist: []string{"ИС-21"}, query: "", want: false},
		{name: "case-sensitive miss", whitelist: []string{"ИС-21"}, query: "ис-21", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := entities.NewTeacherScope(1, tc.whitelist)
			assert.Equal(t, tc.want, s.AllowsGroup(tc.query))
		})
	}
}

func TestTeacherScope_AllowsGroupPtr(t *testing.T) {
	allowed := "ИС-21"
	denied := "ИС-99"

	tests := []struct {
		name      string
		whitelist []string
		query     *string
		want      bool
	}{
		{name: "nil pointer denied (cannot affirm membership)", whitelist: []string{allowed}, query: nil, want: false},
		{name: "match through pointer", whitelist: []string{allowed}, query: &allowed, want: true},
		{name: "no match through pointer", whitelist: []string{allowed}, query: &denied, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := entities.NewTeacherScope(1, tc.whitelist)
			assert.Equal(t, tc.want, s.AllowsGroupPtr(tc.query))
		})
	}
}

func TestTeacherScope_FilterGroupNames(t *testing.T) {
	tests := []struct {
		name      string
		whitelist []string
		input     []string
		want      []string
	}{
		{
			name:      "filters out non-allowed names",
			whitelist: []string{"ИС-21", "ИС-22"},
			input:     []string{"ИС-21", "ИС-99", "ИС-22", "ПИ-31"},
			want:      []string{"ИС-21", "ИС-22"},
		},
		{
			name:      "preserves input order",
			whitelist: []string{"A", "B", "C"},
			input:     []string{"C", "X", "A", "Y", "B"},
			want:      []string{"C", "A", "B"},
		},
		{
			name:      "empty whitelist returns empty",
			whitelist: nil,
			input:     []string{"X", "Y"},
			want:      []string{},
		},
		{
			name:      "empty input returns empty",
			whitelist: []string{"A"},
			input:     []string{},
			want:      []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := entities.NewTeacherScope(1, tc.whitelist)
			got := s.FilterGroupNames(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTeacherScope_AllowedGroupNames(t *testing.T) {
	tests := []struct {
		name      string
		whitelist []string
		wantSet   map[string]struct{}
	}{
		{
			name:      "returns deduplicated whitelist",
			whitelist: []string{"ИС-21", "ИС-22", "ИС-21"},
			wantSet:   map[string]struct{}{"ИС-21": {}, "ИС-22": {}},
		},
		{
			name:      "drops empty strings",
			whitelist: []string{"ИС-21", "", "ПИ-31"},
			wantSet:   map[string]struct{}{"ИС-21": {}, "ПИ-31": {}},
		},
		{
			name:      "empty for nil whitelist",
			whitelist: nil,
			wantSet:   map[string]struct{}{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := entities.NewTeacherScope(1, tc.whitelist)
			got := s.AllowedGroupNames()
			gotSet := make(map[string]struct{}, len(got))
			for _, n := range got {
				gotSet[n] = struct{}{}
			}
			assert.Equal(t, tc.wantSet, gotSet)
			assert.Len(t, got, len(tc.wantSet), "result must be deduplicated")
		})
	}
}

func TestErrAnalyticsScopeForbidden_IsExportedSentinel(t *testing.T) {
	require.NotNil(t, entities.ErrAnalyticsScopeForbidden, "sentinel must exist for errors.Is matching")
	assert.NotEmpty(t, entities.ErrAnalyticsScopeForbidden.Error())
}
