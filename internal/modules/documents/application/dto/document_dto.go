// Package dto contains Data Transfer Objects for the documents module.
package dto

import (
	"mime/multipart"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// CreateDocumentInput represents input for creating a new document
type CreateDocumentInput struct {
	Title          string                `json:"title" validate:"required,min=1,max=500"`
	DocumentTypeID int64                 `json:"document_type_id" validate:"required"`
	CategoryID     *int64                `json:"category_id,omitempty"`
	Subject        *string               `json:"subject,omitempty"`
	Content        *string               `json:"content,omitempty"`
	RecipientID    *int64                `json:"recipient_id,omitempty"`
	Deadline       *time.Time            `json:"deadline,omitempty"`
	Importance     *string               `json:"importance,omitempty" validate:"omitempty,oneof=low normal high urgent"`
	IsPublic       bool                  `json:"is_public"`
	File           *multipart.FileHeader `json:"-"` // from form
}

// UpdateDocumentInput represents input for updating a document
type UpdateDocumentInput struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Subject     *string    `json:"subject,omitempty"`
	Content     *string    `json:"content,omitempty"`
	FileName    *string    `json:"file_name,omitempty" validate:"omitempty,max=500"`
	CategoryID  *int64     `json:"category_id,omitempty"`
	RecipientID *int64     `json:"recipient_id,omitempty"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Importance  *string    `json:"importance,omitempty" validate:"omitempty,oneof=low normal high urgent"`
	IsPublic    *bool      `json:"is_public,omitempty"`
}

// UploadFileInput represents input for uploading a file to a document
type UploadFileInput struct {
	DocumentID int64                 `json:"-"`
	File       *multipart.FileHeader `json:"-" validate:"required"`
}

// DocumentOutput represents output for a single document
type DocumentOutput struct {
	ID                 int64                  `json:"id"`
	DocumentTypeID     int64                  `json:"document_type_id"`
	DocumentTypeName   string                 `json:"document_type_name,omitempty"`
	CategoryID         *int64                 `json:"category_id,omitempty"`
	CategoryName       *string                `json:"category_name,omitempty"`
	RegistrationNumber *string                `json:"registration_number,omitempty"`
	RegistrationDate   *time.Time             `json:"registration_date,omitempty"`
	Title              string                 `json:"title"`
	Subject            *string                `json:"subject,omitempty"`
	Content            *string                `json:"content,omitempty"`
	AuthorID           int64                  `json:"author_id"`
	AuthorName         string                 `json:"author_name,omitempty"`
	RecipientID        *int64                 `json:"recipient_id,omitempty"`
	RecipientName      *string                `json:"recipient_name,omitempty"`
	Status             string                 `json:"status"`
	FileName           *string                `json:"file_name,omitempty"`
	FileSize           *int64                 `json:"file_size,omitempty"`
	MimeType           *string                `json:"mime_type,omitempty"`
	HasFile            bool                   `json:"has_file"`
	Version            int                    `json:"version"`
	Deadline           *time.Time             `json:"deadline,omitempty"`
	ExecutionDate      *time.Time             `json:"execution_date,omitempty"`
	Importance         string                 `json:"importance"`
	IsPublic           bool                   `json:"is_public"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// DocumentListOutput represents paginated list of documents
type DocumentListOutput struct {
	Documents  []*DocumentOutput `json:"documents"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// DocumentFilterInput represents filter options for listing documents
type DocumentFilterInput struct {
	AuthorID       *int64  `form:"author_id"`
	RecipientID    *int64  `form:"recipient_id"`
	DocumentTypeID *int64  `form:"document_type_id"`
	CategoryID     *int64  `form:"category_id"`
	Status         *string `form:"status"`
	Importance     *string `form:"importance"`
	IsPublic       *bool   `form:"is_public"`
	Search         *string `form:"search"`
	FromDate       *string `form:"from_date"`
	ToDate         *string `form:"to_date"`
	Page           int     `form:"page,default=1"`
	PageSize       int     `form:"page_size,default=20"`
	OrderBy        *string `form:"order_by"`
	// Access control fields (set by handler, not from form)
	CurrentUserID   int64  `form:"-"`
	CurrentUserRole string `form:"-"`
}

// ToDocumentOutput converts entity to output DTO
func ToDocumentOutput(doc *entities.Document) *DocumentOutput {
	output := &DocumentOutput{
		ID:                 doc.ID,
		DocumentTypeID:     doc.DocumentTypeID,
		CategoryID:         doc.CategoryID,
		RegistrationNumber: doc.RegistrationNumber,
		RegistrationDate:   doc.RegistrationDate,
		Title:              doc.Title,
		Subject:            doc.Subject,
		Content:            doc.Content,
		AuthorID:           doc.AuthorID,
		RecipientID:        doc.RecipientID,
		Status:             string(doc.Status),
		FileName:           doc.FileName,
		FileSize:           doc.FileSize,
		MimeType:           doc.MimeType,
		HasFile:            doc.HasFile(),
		Version:            doc.Version,
		Deadline:           doc.Deadline,
		ExecutionDate:      doc.ExecutionDate,
		Importance:         string(doc.Importance),
		IsPublic:           doc.IsPublic,
		Metadata:           doc.Metadata,
		CreatedAt:          doc.CreatedAt,
		UpdatedAt:          doc.UpdatedAt,
	}
	// Add author and recipient names if populated
	if doc.AuthorName != nil {
		output.AuthorName = *doc.AuthorName
	}
	if doc.RecipientName != nil {
		output.RecipientName = doc.RecipientName
	}
	return output
}

// FileDownloadOutput represents output for file download
type FileDownloadOutput struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// DocumentTypeOutput represents output for document type
type DocumentTypeOutput struct {
	ID                   int64   `json:"id"`
	Name                 string  `json:"name"`
	Code                 string  `json:"code"`
	Description          *string `json:"description,omitempty"`
	RequiresApproval     bool    `json:"requires_approval"`
	RequiresRegistration bool    `json:"requires_registration"`
}

// DocumentCategoryOutput represents output for document category
type DocumentCategoryOutput struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ParentID    *int64  `json:"parent_id,omitempty"`
}

// SearchInput represents input for full-text search
type SearchInput struct {
	Query          string  `form:"q" validate:"required,min=1,max=500"`
	DocumentTypeID *int64  `form:"document_type_id"`
	CategoryID     *int64  `form:"category_id"`
	AuthorID       *int64  `form:"author_id"`
	Status         *string `form:"status"`
	Importance     *string `form:"importance"`
	FromDate       *string `form:"from_date"`
	ToDate         *string `form:"to_date"`
	Page           int     `form:"page,default=1"`
	PageSize       int     `form:"page_size,default=20"`
	// Access control fields (set by handler, not from form)
	CurrentUserID   int64  `form:"-"`
	CurrentUserRole string `form:"-"`
}

// SearchResultOutput represents a single search result with highlighted matches
type SearchResultOutput struct {
	Document           *DocumentOutput `json:"document"`
	Rank               float64         `json:"rank"`
	HighlightedTitle   string          `json:"highlighted_title"`
	HighlightedSubject string          `json:"highlighted_subject"`
	HighlightedContent string          `json:"highlighted_content"`
}

// SearchOutput represents paginated search results
type SearchOutput struct {
	Results    []*SearchResultOutput `json:"results"`
	Query      string                `json:"query"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}
