// Package entities contains domain entities for the documents module.
package entities

import (
	"errors"
	"fmt"
	"time"
)

// ErrDocumentEditDenied is returned by CanBeEditedBy when the caller
// is not allowed to mutate the document (wrong role, or teacher
// trying to edit another author's document). Exposed as a sentinel so
// handlers can errors.Is it and map to a stable 403 response without
// string parsing.
var ErrDocumentEditDenied = errors.New("not allowed to edit this document")

// ErrCannotSubmit signals Submit invoked on a non-draft document
// (already submitted, approved, rejected, etc.). v0.148.0 workflow gate.
//
// Issue: #227
var ErrCannotSubmit = errors.New("document: cannot submit, status must be draft")

// ErrCannotApprove signals Approve invoked on a document not in the
// approval queue (e.g. caller pressed approve on a draft or already
// approved document).
//
// Issue: #227
var ErrCannotApprove = errors.New("document: cannot approve, status must be approval")

// ErrCannotReject signals Reject invoked on a document not in the
// approval queue.
//
// Issue: #227
var ErrCannotReject = errors.New("document: cannot reject, status must be approval")

// ErrCannotRegister signals Register invoked on a non-approved
// document. Phase 2 workflow gate.
//
// Issue: #230
var ErrCannotRegister = errors.New("document: cannot register, status must be approved")

// ErrInvalidRegistrationNumber signals empty / whitespace-only / too
// short registration number passed к Register. Min length 3 после trim.
//
// Issue: #230
var ErrInvalidRegistrationNumber = errors.New("document: registration number invalid (must be ≥3 chars after trim)")

// DocumentStatus represents the status of a document in workflow
type DocumentStatus string

// DocumentStatus values.
const (
	DocumentStatusDraft      DocumentStatus = "draft"
	DocumentStatusRegistered DocumentStatus = "registered"
	DocumentStatusRouting    DocumentStatus = "routing"
	DocumentStatusApproval   DocumentStatus = "approval"
	DocumentStatusApproved   DocumentStatus = "approved"
	DocumentStatusRejected   DocumentStatus = "rejected"
	DocumentStatusExecution  DocumentStatus = "execution"
	DocumentStatusExecuted   DocumentStatus = "executed"
	DocumentStatusArchived   DocumentStatus = "archived"
)

// DocumentImportance represents the importance level of a document
type DocumentImportance string

// DocumentImportance values.
const (
	ImportanceLow    DocumentImportance = "low"
	ImportanceNormal DocumentImportance = "normal"
	ImportanceHigh   DocumentImportance = "high"
	ImportanceUrgent DocumentImportance = "urgent"
)

// Document represents a document entity in the documents domain
type Document struct {
	ID             int64  `json:"id"`
	DocumentTypeID int64  `json:"document_type_id"`
	CategoryID     *int64 `json:"category_id,omitempty"`

	// Registration data
	RegistrationNumber *string    `json:"registration_number,omitempty"`
	RegistrationDate   *time.Time `json:"registration_date,omitempty"`

	// Main information
	Title   string  `json:"title"`
	Subject *string `json:"subject,omitempty"`
	Content *string `json:"content,omitempty"`

	// Author details
	AuthorID         int64   `json:"author_id"`
	AuthorName       *string `json:"author_name,omitempty"` // Populated via JOIN
	AuthorDepartment *string `json:"author_department,omitempty"`
	AuthorPosition   *string `json:"author_position,omitempty"`

	// Recipient details
	RecipientID         *int64  `json:"recipient_id,omitempty"`
	RecipientName       *string `json:"recipient_name,omitempty"` // Populated via JOIN
	RecipientDepartment *string `json:"recipient_department,omitempty"`
	RecipientPosition   *string `json:"recipient_position,omitempty"`
	RecipientExternal   *string `json:"recipient_external,omitempty"`

	// Status and workflow
	Status DocumentStatus `json:"status"`

	// File information
	FileName *string `json:"file_name,omitempty"`
	FilePath *string `json:"file_path,omitempty"`
	FileSize *int64  `json:"file_size,omitempty"`
	MimeType *string `json:"mime_type,omitempty"`

	// Versioning
	Version          int    `json:"version"`
	ParentDocumentID *int64 `json:"parent_document_id,omitempty"`

	// Deadlines
	Deadline      *time.Time `json:"deadline,omitempty"`
	ExecutionDate *time.Time `json:"execution_date,omitempty"`

	// Metadata
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	IsPublic   bool                   `json:"is_public"`
	Importance DocumentImportance     `json:"importance"`

	// Audit
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// Workflow audit trail (v0.148.0 — issue #227). Nullable so
	// pre-v0.148.0 documents (which never traversed approval gates)
	// keep clean JSON output. Mirror к curriculum's approved_by /
	// _at + adds rejected/submitted columns.
	SubmittedBy    *int64     `json:"submitted_by,omitempty"`
	SubmittedAt    *time.Time `json:"submitted_at,omitempty"`
	ApprovedBy     *int64     `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectedBy     *int64     `json:"rejected_by,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`
	RejectedReason *string    `json:"rejected_reason,omitempty"`
	// v0.149.0 Phase 2 — Register transition (#230).
	RegisteredBy *int64 `json:"registered_by,omitempty"`
}

// NewDocument creates a new document with default values
func NewDocument(title string, documentTypeID, authorID int64) *Document {
	now := time.Now()
	return &Document{
		DocumentTypeID: documentTypeID,
		Title:          title,
		AuthorID:       authorID,
		Status:         DocumentStatusDraft,
		Version:        1,
		IsPublic:       false,
		Importance:     ImportanceNormal,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// SetFile sets file information for the document
func (d *Document) SetFile(fileName, filePath, mimeType string, fileSize int64) {
	d.FileName = &fileName
	d.FilePath = &filePath
	d.MimeType = &mimeType
	d.FileSize = &fileSize
	d.UpdatedAt = time.Now()
}

// ClearFile removes file information from the document
func (d *Document) ClearFile() {
	d.FileName = nil
	d.FilePath = nil
	d.MimeType = nil
	d.FileSize = nil
	d.UpdatedAt = time.Now()
}

// Register registers the document с the number + date + audit trail.
// v0.149.0 Phase 2 (#230) — extends original Register с registrarID +
// now params + invariant check (status must be approved) + non-empty
// number validation. RED stub returns ErrCannotRegister
// unconditionally; GREEN replaces the body.
//
// Issue: #230
func (d *Document) Register(registrationNumber string, registrarID int64, now time.Time) error {
	return ErrCannotRegister
}

// IsDraft checks if document is in draft status
func (d *Document) IsDraft() bool {
	return d.Status == DocumentStatusDraft
}

// IsDeleted checks if document is soft-deleted
func (d *Document) IsDeleted() bool {
	return d.DeletedAt != nil
}

// SoftDelete marks the document as deleted
func (d *Document) SoftDelete() {
	now := time.Now()
	d.DeletedAt = &now
	d.UpdatedAt = now
}

// Restore restores a soft-deleted document
func (d *Document) Restore() {
	d.DeletedAt = nil
	d.UpdatedAt = time.Now()
}

// HasFile checks if document has an attached file
func (d *Document) HasFile() bool {
	return d.FilePath != nil && *d.FilePath != ""
}

// Submit moves a draft document into the approval queue. Sets the
// SubmittedBy + SubmittedAt audit fields. Returns ErrCannotSubmit
// when the current status is not Draft — workflow invariant guarded
// at the domain boundary.
//
// Issue: #227
func (d *Document) Submit(actorID int64, now time.Time) error {
	if d.Status != DocumentStatusDraft {
		return fmt.Errorf("%w: status %q", ErrCannotSubmit, string(d.Status))
	}
	d.Status = DocumentStatusApproval
	d.SubmittedBy = &actorID
	d.SubmittedAt = &now
	d.UpdatedAt = now
	return nil
}

// Approve advances an approval-queue document к the approved state.
// Sets ApprovedBy + ApprovedAt audit fields. Returns ErrCannotApprove
// when the current status is not Approval.
//
// Issue: #227
func (d *Document) Approve(adminID int64, now time.Time) error {
	if d.Status != DocumentStatusApproval {
		return fmt.Errorf("%w: status %q", ErrCannotApprove, string(d.Status))
	}
	d.Status = DocumentStatusApproved
	d.ApprovedBy = &adminID
	d.ApprovedAt = &now
	d.UpdatedAt = now
	return nil
}

// Reject marks an approval-queue document as rejected с обоснованием.
// Sets RejectedBy + RejectedAt + RejectedReason audit fields. Returns
// ErrCannotReject when the current status is not Approval, or
// ErrRejectionReasonInvalid when the reason VO is zero-value.
//
// Issue: #227
func (d *Document) Reject(adminID int64, reason RejectionReason, now time.Time) error {
	if reason.IsZero() {
		return fmt.Errorf("%w: zero-value reason", ErrRejectionReasonInvalid)
	}
	if d.Status != DocumentStatusApproval {
		return fmt.Errorf("%w: status %q", ErrCannotReject, string(d.Status))
	}
	d.Status = DocumentStatusRejected
	d.RejectedBy = &adminID
	d.RejectedAt = &now
	reasonStr := reason.String()
	d.RejectedReason = &reasonStr
	d.UpdatedAt = now
	return nil
}

// CanBeEditedBy reports whether a user holding the given role is
// allowed to mutate this document.
//
// The rule encodes the audit-report decision:
//   - methodist / academic_secretary / system_admin: any document;
//   - teacher: only own (userID == AuthorID);
//   - student / unknown role: never — defense-in-depth alongside the
//     handler-level RequireNonStudent middleware.
//
// Returns nil on allow, ErrDocumentEditDenied on deny.
func (d *Document) CanBeEditedBy(userID int64, role UserRole) error {
	switch role {
	case RoleMethodist, RoleAcademicSecretary, RoleSystemAdmin:
		return nil
	case RoleTeacher:
		if userID == d.AuthorID {
			return nil
		}
		return ErrDocumentEditDenied
	default:
		return ErrDocumentEditDenied
	}
}
