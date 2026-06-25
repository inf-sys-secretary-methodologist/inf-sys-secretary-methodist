package handlers

import (
	"time"

	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// AttemptDTO is the JSON projection of a resit attempt.
type AttemptDTO struct {
	ID            int64   `json:"id"`
	AttemptNo     int     `json:"attempt_no"`
	IsCommission  bool    `json:"is_commission"`
	ScheduledDate string  `json:"scheduled_date"`
	Examiner      string  `json:"examiner"`
	Result        string  `json:"result"`
	Grade         *int    `json:"grade,omitempty"`
	RecordedBy    *int64  `json:"recorded_by,omitempty"`
	RecordedAt    *string `json:"recorded_at,omitempty"`
}

// DebtDTO is the full JSON projection of a StudentDebt aggregate (root +
// attempt timeline), returned by GET /:id.
type DebtDTO struct {
	ID              int64        `json:"id"`
	StudentFullName string       `json:"student_full_name"`
	GroupName       string       `json:"group_name"`
	DisciplineName  string       `json:"discipline_name"`
	Semester        int          `json:"semester"`
	ControlForm     string       `json:"control_form"`
	StudentUserID   *int64       `json:"student_user_id,omitempty"`
	DisciplineID    *int64       `json:"discipline_id,omitempty"`
	SourceRef       string       `json:"source_ref,omitempty"`
	Status          string       `json:"status"`
	Version         int          `json:"version"`
	CreatedAt       string       `json:"created_at"`
	UpdatedAt       string       `json:"updated_at"`
	Attempts        []AttemptDTO `json:"attempts"`
}

// DebtListItemDTO is the lightweight JSON projection for list endpoints
// (root-only, no attempts).
type DebtListItemDTO struct {
	ID              int64  `json:"id"`
	StudentFullName string `json:"student_full_name"`
	GroupName       string `json:"group_name"`
	DisciplineName  string `json:"discipline_name"`
	Semester        int    `json:"semester"`
	ControlForm     string `json:"control_form"`
	StudentUserID   *int64 `json:"student_user_id,omitempty"`
	Status          string `json:"status"`
	Version         int    `json:"version"`
}

// DebtListResponse bundles a page of debts with the total match count.
type DebtListResponse struct {
	Items []DebtListItemDTO `json:"items"`
	Total int               `json:"total"`
}

// DebtStatsDTO is the JSON projection of the dashboard aggregate.
type DebtStatsDTO struct {
	Total          int `json:"total"`
	Open           int `json:"open"`
	ResitScheduled int `json:"resit_scheduled"`
	Commission     int `json:"commission"`
	ClosedPassed   int `json:"closed_passed"`
	ClosedFailed   int `json:"closed_failed"`
}

// ImportRowErrorDTO is the JSON projection of a single failed source row.
type ImportRowErrorDTO struct {
	Row      int    `json:"row"`
	Identity string `json:"identity"`
	Message  string `json:"message"`
}

// ImportResultDTO is the JSON projection of the import log: how many rows
// were created / updated / skipped, plus per-row errors. Errors is always a
// (possibly empty) array so the frontend can map over it unconditionally.
type ImportResultDTO struct {
	Created int                 `json:"created"`
	Updated int                 `json:"updated"`
	Skipped int                 `json:"skipped"`
	Errors  []ImportRowErrorDTO `json:"errors"`
}

func mapImportResult(res sdUsecases.ImportResult) ImportResultDTO {
	errs := make([]ImportRowErrorDTO, 0, len(res.Errors))
	for _, e := range res.Errors {
		errs = append(errs, ImportRowErrorDTO{Row: e.Row, Identity: e.Identity, Message: e.Message})
	}
	return ImportResultDTO{
		Created: res.Created,
		Updated: res.Updated,
		Skipped: res.Skipped,
		Errors:  errs,
	}
}

func formatTime(t time.Time) string { return t.Format(time.RFC3339) }

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func mapAttempt(a *entities.ResitAttempt) AttemptDTO {
	return AttemptDTO{
		ID:            a.ID,
		AttemptNo:     a.AttemptNo,
		IsCommission:  a.IsCommission,
		ScheduledDate: formatTime(a.ScheduledDate()),
		Examiner:      a.Examiner(),
		Result:        string(a.Result()),
		Grade:         a.Grade(),
		RecordedBy:    a.RecordedBy(),
		RecordedAt:    formatTimePtr(a.RecordedAt()),
	}
}

func mapDebt(d *entities.StudentDebt) DebtDTO {
	attempts := d.Attempts()
	dtos := make([]AttemptDTO, 0, len(attempts))
	for _, a := range attempts {
		dtos = append(dtos, mapAttempt(a))
	}
	return DebtDTO{
		ID:              d.ID,
		StudentFullName: d.StudentFullName,
		GroupName:       d.GroupName,
		DisciplineName:  d.DisciplineName,
		Semester:        d.Semester,
		ControlForm:     string(d.ControlForm),
		StudentUserID:   d.StudentUserID,
		DisciplineID:    d.DisciplineID,
		SourceRef:       d.SourceRef,
		Status:          string(d.Status()),
		Version:         d.Version,
		CreatedAt:       formatTime(d.CreatedAt()),
		UpdatedAt:       formatTime(d.UpdatedAt()),
		Attempts:        dtos,
	}
}

func mapListItem(it repositories.StudentDebtListItem) DebtListItemDTO {
	return DebtListItemDTO{
		ID:              it.ID,
		StudentFullName: it.StudentFullName,
		GroupName:       it.GroupName,
		DisciplineName:  it.DisciplineName,
		Semester:        it.Semester,
		ControlForm:     string(it.ControlForm),
		StudentUserID:   it.StudentUserID,
		Status:          string(it.Status),
		Version:         it.Version,
	}
}

func mapList(res repositories.StudentDebtListResult) DebtListResponse {
	items := make([]DebtListItemDTO, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, mapListItem(it))
	}
	return DebtListResponse{Items: items, Total: res.Total}
}

func mapStats(s repositories.StudentDebtStats) DebtStatsDTO {
	return DebtStatsDTO{
		Total:          s.Total,
		Open:           s.Open,
		ResitScheduled: s.ResitScheduled,
		Commission:     s.Commission,
		ClosedPassed:   s.ClosedPassed,
		ClosedFailed:   s.ClosedFailed,
	}
}
