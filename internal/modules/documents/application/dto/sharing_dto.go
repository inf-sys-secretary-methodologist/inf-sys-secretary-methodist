package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// ShareDocumentInput represents input for sharing a document with a user
type ShareDocumentInput struct {
	DocumentID int64  `json:"-"`
	UserID     *int64 `json:"user_id,omitempty" validate:"required_without=Role"`
	Role       *string `json:"role,omitempty" validate:"required_without=UserID,omitempty,oneof=admin secretary methodist teacher student"`
	Permission string `json:"permission" validate:"required,oneof=read write delete admin"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// UpdatePermissionInput represents input for updating a permission
type UpdatePermissionInput struct {
	Permission string `json:"permission" validate:"required,oneof=read write delete admin"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// PermissionOutput represents output for a document permission
type PermissionOutput struct {
	ID            int64      `json:"id"`
	DocumentID    int64      `json:"document_id"`
	UserID        *int64     `json:"user_id,omitempty"`
	UserName      *string    `json:"user_name,omitempty"`
	UserEmail     *string    `json:"user_email,omitempty"`
	Role          *string    `json:"role,omitempty"`
	Permission    string     `json:"permission"`
	GrantedBy     *int64     `json:"granted_by,omitempty"`
	GrantedByName *string    `json:"granted_by_name,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ToPermissionOutput converts an entity to DTO
func ToPermissionOutput(p *entities.DocumentPermission) *PermissionOutput {
	var role *string
	if p.Role != nil {
		r := string(*p.Role)
		role = &r
	}
	return &PermissionOutput{
		ID:            p.ID,
		DocumentID:    p.DocumentID,
		UserID:        p.UserID,
		UserName:      p.UserName,
		UserEmail:     p.UserEmail,
		Role:          role,
		Permission:    string(p.Permission),
		GrantedBy:     p.GrantedBy,
		GrantedByName: p.GrantedByName,
		ExpiresAt:     p.ExpiresAt,
		CreatedAt:     p.CreatedAt,
	}
}

// ToPermissionOutputList converts a slice of entities to DTOs
func ToPermissionOutputList(permissions []*entities.DocumentPermission) []*PermissionOutput {
	result := make([]*PermissionOutput, len(permissions))
	for i, p := range permissions {
		result[i] = ToPermissionOutput(p)
	}
	return result
}

// CreatePublicLinkInput represents input for creating a public link
type CreatePublicLinkInput struct {
	DocumentID int64  `json:"-"`
	Permission string `json:"permission" validate:"required,oneof=read download"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	MaxUses    *int   `json:"max_uses,omitempty" validate:"omitempty,min=1"`
	Password   *string `json:"password,omitempty" validate:"omitempty,min=4"`
}

// UpdatePublicLinkInput represents input for updating a public link
type UpdatePublicLinkInput struct {
	Permission *string `json:"permission,omitempty" validate:"omitempty,oneof=read download"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	MaxUses    *int   `json:"max_uses,omitempty" validate:"omitempty,min=1"`
	Password   *string `json:"password,omitempty" validate:"omitempty,min=4"`
	IsActive   *bool  `json:"is_active,omitempty"`
}

// PublicLinkOutput represents output for a public link
type PublicLinkOutput struct {
	ID            int64      `json:"id"`
	DocumentID    int64      `json:"document_id"`
	DocumentTitle *string    `json:"document_title,omitempty"`
	Token         string     `json:"token"`
	URL           string     `json:"url"` // Full URL for sharing
	Permission    string     `json:"permission"`
	CreatedBy     int64      `json:"created_by"`
	CreatedByName *string    `json:"created_by_name,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	MaxUses       *int       `json:"max_uses,omitempty"`
	UseCount      int        `json:"use_count"`
	HasPassword   bool       `json:"has_password"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ToPublicLinkOutput converts an entity to DTO
func ToPublicLinkOutput(l *entities.PublicLink, baseURL string) *PublicLinkOutput {
	return &PublicLinkOutput{
		ID:            l.ID,
		DocumentID:    l.DocumentID,
		DocumentTitle: l.DocumentTitle,
		Token:         l.Token,
		URL:           baseURL + "/api/v1/public/documents/" + l.Token,
		Permission:    string(l.Permission),
		CreatedBy:     l.CreatedBy,
		CreatedByName: l.CreatedByName,
		ExpiresAt:     l.ExpiresAt,
		MaxUses:       l.MaxUses,
		UseCount:      l.UseCount,
		HasPassword:   l.HasPassword(),
		IsActive:      l.IsActive,
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
	}
}

// ToPublicLinkOutputList converts a slice of entities to DTOs
func ToPublicLinkOutputList(links []*entities.PublicLink, baseURL string) []*PublicLinkOutput {
	result := make([]*PublicLinkOutput, len(links))
	for i, l := range links {
		result[i] = ToPublicLinkOutput(l, baseURL)
	}
	return result
}

// AccessPublicLinkInput represents input for accessing a public link
type AccessPublicLinkInput struct {
	Token    string  `json:"-"`
	Password *string `json:"password,omitempty"`
}

// DocumentAccessOutput represents the document data accessible via public link
type DocumentAccessOutput struct {
	ID                 int64      `json:"id"`
	Title              string     `json:"title"`
	Subject            *string    `json:"subject,omitempty"`
	Content            *string    `json:"content,omitempty"`
	AuthorName         *string    `json:"author_name,omitempty"`
	RegistrationNumber *string    `json:"registration_number,omitempty"`
	RegistrationDate   *time.Time `json:"registration_date,omitempty"`
	FileName           *string    `json:"file_name,omitempty"`
	FileSize           *int64     `json:"file_size,omitempty"`
	MimeType           *string    `json:"mime_type,omitempty"`
	CanDownload        bool       `json:"can_download"`
	CreatedAt          time.Time  `json:"created_at"`
}

// SharedDocumentsFilter represents filter options for listing shared documents
type SharedDocumentsFilter struct {
	UserID     int64
	UserRole   string
	Permission *string
	Limit      int
	Offset     int
}

// SharedWithInfo represents info about who a document is shared with
type SharedWithInfo struct {
	PermissionID int64      `json:"permission_id"`
	UserID       *int64     `json:"user_id,omitempty"`
	UserName     *string    `json:"user_name,omitempty"`
	UserEmail    *string    `json:"user_email,omitempty"`
	Role         *string    `json:"role,omitempty"`
	Permission   string     `json:"permission"`
	GrantedAt    time.Time  `json:"granted_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// MySharedDocumentOutput represents a document the user has shared with others
type MySharedDocumentOutput struct {
	DocumentID    int64            `json:"document_id"`
	DocumentTitle string           `json:"document_title"`
	SharedWith    []SharedWithInfo `json:"shared_with"`
}
