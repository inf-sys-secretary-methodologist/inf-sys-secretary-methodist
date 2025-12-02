// Package dto contains data transfer objects for the announcements module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
)

// CreateAnnouncementRequest represents a request to create an announcement.
type CreateAnnouncementRequest struct {
	Title          string                      `json:"title" validate:"required,min=1,max=500"`
	Content        string                      `json:"content" validate:"required,min=1"`
	Summary        *string                     `json:"summary,omitempty" validate:"omitempty,max=1000"`
	Priority       domain.AnnouncementPriority `json:"priority" validate:"required"`
	TargetAudience domain.TargetAudience       `json:"target_audience" validate:"required"`
	PublishAt      *time.Time                  `json:"publish_at,omitempty"`
	ExpireAt       *time.Time                  `json:"expire_at,omitempty"`
	IsPinned       bool                        `json:"is_pinned"`
	Tags           []string                    `json:"tags,omitempty"`
}

// UpdateAnnouncementRequest represents a request to update an announcement.
type UpdateAnnouncementRequest struct {
	Title          *string                      `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Content        *string                      `json:"content,omitempty" validate:"omitempty,min=1"`
	Summary        *string                      `json:"summary,omitempty" validate:"omitempty,max=1000"`
	Priority       *domain.AnnouncementPriority `json:"priority,omitempty"`
	TargetAudience *domain.TargetAudience       `json:"target_audience,omitempty"`
	PublishAt      *time.Time                   `json:"publish_at,omitempty"`
	ExpireAt       *time.Time                   `json:"expire_at,omitempty"`
	IsPinned       *bool                        `json:"is_pinned,omitempty"`
	Tags           []string                     `json:"tags,omitempty"`
}

// AnnouncementResponse represents an announcement in API responses.
type AnnouncementResponse struct {
	ID             int64                       `json:"id"`
	Title          string                      `json:"title"`
	Content        string                      `json:"content"`
	Summary        *string                     `json:"summary,omitempty"`
	AuthorID       int64                       `json:"author_id"`
	Author         *AuthorResponse             `json:"author,omitempty"`
	Status         domain.AnnouncementStatus   `json:"status"`
	Priority       domain.AnnouncementPriority `json:"priority"`
	TargetAudience domain.TargetAudience       `json:"target_audience"`
	PublishAt      *time.Time                  `json:"publish_at,omitempty"`
	ExpireAt       *time.Time                  `json:"expire_at,omitempty"`
	IsPinned       bool                        `json:"is_pinned"`
	ViewCount      int64                       `json:"view_count"`
	Tags           []string                    `json:"tags,omitempty"`
	Attachments    []AttachmentResponse        `json:"attachments,omitempty"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

// AuthorResponse represents the author in API responses.
type AuthorResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AttachmentResponse represents an attachment in API responses.
type AttachmentResponse struct {
	ID        int64     `json:"id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

// ListAnnouncementsRequest represents a request to list announcements.
type ListAnnouncementsRequest struct {
	AuthorID       *int64                       `json:"author_id,omitempty"`
	Status         *domain.AnnouncementStatus   `json:"status,omitempty"`
	Priority       *domain.AnnouncementPriority `json:"priority,omitempty"`
	TargetAudience *domain.TargetAudience       `json:"target_audience,omitempty"`
	IsPinned       *bool                        `json:"is_pinned,omitempty"`
	Search         *string                      `json:"search,omitempty"`
	Tags           []string                     `json:"tags,omitempty"`
	Limit          int                          `json:"limit" validate:"min=1,max=100"`
	Offset         int                          `json:"offset" validate:"min=0"`
}

// ListAnnouncementsResponse represents a paginated list response.
type ListAnnouncementsResponse struct {
	Announcements []AnnouncementResponse `json:"announcements"`
	Total         int64                  `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

// ToResponse converts an Announcement entity to AnnouncementResponse.
func ToResponse(a *entities.Announcement) *AnnouncementResponse {
	resp := &AnnouncementResponse{
		ID:             a.ID,
		Title:          a.Title,
		Content:        a.Content,
		Summary:        a.Summary,
		AuthorID:       a.AuthorID,
		Status:         a.Status,
		Priority:       a.Priority,
		TargetAudience: a.TargetAudience,
		PublishAt:      a.PublishAt,
		ExpireAt:       a.ExpireAt,
		IsPinned:       a.IsPinned,
		ViewCount:      a.ViewCount,
		Tags:           a.Tags,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}

	if a.Author != nil {
		resp.Author = &AuthorResponse{
			ID:    a.Author.ID,
			Name:  a.Author.Name,
			Email: a.Author.Email,
		}
	}

	if len(a.Attachments) > 0 {
		resp.Attachments = make([]AttachmentResponse, len(a.Attachments))
		for i, att := range a.Attachments {
			resp.Attachments[i] = AttachmentResponse{
				ID:        att.ID,
				FileName:  att.FileName,
				FileSize:  att.FileSize,
				MimeType:  att.MimeType,
				CreatedAt: att.CreatedAt,
			}
		}
	}

	return resp
}

// ToResponseList converts a slice of Announcement entities to responses.
func ToResponseList(announcements []*entities.Announcement) []AnnouncementResponse {
	responses := make([]AnnouncementResponse, len(announcements))
	for i, a := range announcements {
		responses[i] = *ToResponse(a)
	}
	return responses
}
