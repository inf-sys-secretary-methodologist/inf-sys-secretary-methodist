package entities_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

func validOrderInput() entities.NewMinobrnaukiOrderInput {
	docID := int64(7)
	return entities.NewMinobrnaukiOrderInput{
		OrderNumber: "№ 1078 от 12.05.2026",
		Title:       "Об изменении ФГОС 09.03.01",
		PublishedAt: time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
		DocumentID:  &docID,
		ChangeScope: domain.MinobrnaukiOrderChangeScopeMajor,
		Summary:     "Обновлён перечень компетенций",
		UploadedBy:  42,
	}
}

func TestNewMinobrnaukiOrder_HappyPath(t *testing.T) {
	o, err := entities.NewMinobrnaukiOrder(validOrderInput())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if o.OrderNumber() != "№ 1078 от 12.05.2026" {
		t.Errorf("order_number = %q", o.OrderNumber())
	}
	if o.Title() != "Об изменении ФГОС 09.03.01" {
		t.Errorf("title = %q", o.Title())
	}
	if o.ChangeScope() != domain.MinobrnaukiOrderChangeScopeMajor {
		t.Errorf("change_scope = %q", o.ChangeScope())
	}
	if o.DocumentID() == nil || *o.DocumentID() != 7 {
		t.Errorf("document_id = %v", o.DocumentID())
	}
	if o.UploadedBy() != 42 {
		t.Errorf("uploaded_by = %d", o.UploadedBy())
	}
	if o.CreatedAt().IsZero() {
		t.Error("created_at must be set")
	}
}

func TestNewMinobrnaukiOrder_NilDocumentIDAllowed(t *testing.T) {
	in := validOrderInput()
	in.DocumentID = nil
	o, err := entities.NewMinobrnaukiOrder(in)
	if err != nil {
		t.Fatalf("nil document_id must be allowed, got %v", err)
	}
	if o.DocumentID() != nil {
		t.Errorf("document_id should stay nil, got %v", o.DocumentID())
	}
}

func TestNewMinobrnaukiOrder_TrimsFields(t *testing.T) {
	in := validOrderInput()
	in.OrderNumber = "  № 5  "
	in.Title = "  Заголовок  "
	in.Summary = "  Краткое описание  "
	o, err := entities.NewMinobrnaukiOrder(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.OrderNumber() != "№ 5" {
		t.Errorf("order_number not trimmed: %q", o.OrderNumber())
	}
	if o.Title() != "Заголовок" {
		t.Errorf("title not trimmed: %q", o.Title())
	}
	if o.Summary() != "Краткое описание" {
		t.Errorf("summary not trimmed: %q", o.Summary())
	}
}

func TestNewMinobrnaukiOrder_Invariants(t *testing.T) {
	zeroDoc := int64(0)
	cases := []struct {
		name   string
		mutate func(*entities.NewMinobrnaukiOrderInput)
	}{
		{"empty order_number", func(in *entities.NewMinobrnaukiOrderInput) { in.OrderNumber = "   " }},
		{"over-long order_number", func(in *entities.NewMinobrnaukiOrderInput) { in.OrderNumber = strings.Repeat("x", 101) }},
		{"empty title", func(in *entities.NewMinobrnaukiOrderInput) { in.Title = "" }},
		{"over-long title", func(in *entities.NewMinobrnaukiOrderInput) { in.Title = strings.Repeat("y", 1025) }},
		{"zero published_at", func(in *entities.NewMinobrnaukiOrderInput) { in.PublishedAt = time.Time{} }},
		{"invalid change_scope", func(in *entities.NewMinobrnaukiOrderInput) { in.ChangeScope = "bogus" }},
		{"over-long summary", func(in *entities.NewMinobrnaukiOrderInput) { in.Summary = strings.Repeat("z", 4097) }},
		{"non-positive uploaded_by", func(in *entities.NewMinobrnaukiOrderInput) { in.UploadedBy = 0 }},
		{"non-positive document_id when set", func(in *entities.NewMinobrnaukiOrderInput) { in.DocumentID = &zeroDoc }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validOrderInput()
			tc.mutate(&in)
			_, err := entities.NewMinobrnaukiOrder(in)
			if !errors.Is(err, domain.ErrInvalidMinobrnaukiOrder) {
				t.Errorf("expected ErrInvalidMinobrnaukiOrder, got %v", err)
			}
		})
	}
}

func TestNewMinobrnaukiOrder_EmptySummaryAllowed(t *testing.T) {
	in := validOrderInput()
	in.Summary = ""
	if _, err := entities.NewMinobrnaukiOrder(in); err != nil {
		t.Fatalf("empty summary must be allowed, got %v", err)
	}
}

func TestReconstituteMinobrnaukiOrder(t *testing.T) {
	docID := int64(9)
	created := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	o := entities.ReconstituteMinobrnaukiOrder(entities.ReconstituteMinobrnaukiOrderInput{
		ID:          5,
		OrderNumber: "№ 1",
		Title:       "T",
		PublishedAt: time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
		DocumentID:  &docID,
		ChangeScope: domain.MinobrnaukiOrderChangeScopeMinor,
		Summary:     "S",
		UploadedBy:  3,
		CreatedAt:   created,
	})
	if o.ID() != 5 {
		t.Errorf("id = %d", o.ID())
	}
	if o.ChangeScope() != domain.MinobrnaukiOrderChangeScopeMinor {
		t.Errorf("change_scope = %q", o.ChangeScope())
	}
	if !o.CreatedAt().Equal(created) {
		t.Errorf("created_at = %v", o.CreatedAt())
	}
}

func TestMinobrnaukiOrderChangeScope_IsValid(t *testing.T) {
	cases := []struct {
		scope domain.MinobrnaukiOrderChangeScope
		want  bool
	}{
		{domain.MinobrnaukiOrderChangeScopeMinor, true},
		{domain.MinobrnaukiOrderChangeScopeMajor, true},
		{"", false},
		{"bogus", false},
	}
	for _, tc := range cases {
		if got := tc.scope.IsValid(); got != tc.want {
			t.Errorf("IsValid(%q) = %v, want %v", tc.scope, got, tc.want)
		}
	}
}

func TestMinobrnaukiOrderChangeScope_String(t *testing.T) {
	if domain.MinobrnaukiOrderChangeScopeMajor.String() != "major" {
		t.Errorf("String() = %q", domain.MinobrnaukiOrderChangeScopeMajor.String())
	}
}
