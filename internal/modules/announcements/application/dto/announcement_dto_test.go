package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToResponse_Basic(t *testing.T) {
	now := time.Now()
	summary := "A summary"
	a := &entities.Announcement{
		ID:             1,
		Title:          "Test",
		Content:        "Content",
		Summary:        &summary,
		AuthorID:       42,
		Status:         domain.AnnouncementStatusDraft,
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
		IsPinned:       true,
		ViewCount:      10,
		Tags:           []string{"tag1", "tag2"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	resp := ToResponse(a)

	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, "Test", resp.Title)
	assert.Equal(t, "Content", resp.Content)
	assert.Equal(t, &summary, resp.Summary)
	assert.Equal(t, int64(42), resp.AuthorID)
	assert.Equal(t, domain.AnnouncementStatusDraft, resp.Status)
	assert.Equal(t, domain.AnnouncementPriorityNormal, resp.Priority)
	assert.Equal(t, domain.TargetAudienceAll, resp.TargetAudience)
	assert.True(t, resp.IsPinned)
	assert.Equal(t, int64(10), resp.ViewCount)
	assert.Equal(t, []string{"tag1", "tag2"}, resp.Tags)
	assert.Nil(t, resp.Author)
	assert.Empty(t, resp.Attachments)
}

func TestToResponse_WithAuthor(t *testing.T) {
	a := &entities.Announcement{
		ID:       1,
		Title:    "Test",
		Content:  "Content",
		AuthorID: 42,
		Author: &entities.AnnouncementAuthor{
			ID:    42,
			Name:  "John Doe",
			Email: "john@example.com",
		},
		Status:         domain.AnnouncementStatusPublished,
		Priority:       domain.AnnouncementPriorityHigh,
		TargetAudience: domain.TargetAudienceAll,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	resp := ToResponse(a)

	require.NotNil(t, resp.Author)
	assert.Equal(t, int64(42), resp.Author.ID)
	assert.Equal(t, "John Doe", resp.Author.Name)
	assert.Equal(t, "john@example.com", resp.Author.Email)
}

func TestToResponse_WithAttachments(t *testing.T) {
	now := time.Now()
	a := &entities.Announcement{
		ID:       1,
		Title:    "Test",
		Content:  "Content",
		AuthorID: 1,
		Status:   domain.AnnouncementStatusDraft,
		Priority: domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
		Attachments: []entities.AnnouncementAttachment{
			{
				ID:        10,
				FileName:  "file.pdf",
				FileSize:  1024,
				MimeType:  "application/pdf",
				CreatedAt: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := ToResponse(a)

	require.Len(t, resp.Attachments, 1)
	assert.Equal(t, int64(10), resp.Attachments[0].ID)
	assert.Equal(t, "file.pdf", resp.Attachments[0].FileName)
	assert.Equal(t, int64(1024), resp.Attachments[0].FileSize)
	assert.Equal(t, "application/pdf", resp.Attachments[0].MimeType)
}

func TestToResponseList(t *testing.T) {
	now := time.Now()
	announcements := []*entities.Announcement{
		{
			ID: 1, Title: "First", Content: "C1", AuthorID: 1,
			Status: domain.AnnouncementStatusDraft, Priority: domain.AnnouncementPriorityNormal,
			TargetAudience: domain.TargetAudienceAll, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: 2, Title: "Second", Content: "C2", AuthorID: 2,
			Status: domain.AnnouncementStatusPublished, Priority: domain.AnnouncementPriorityHigh,
			TargetAudience: domain.TargetAudienceAll, CreatedAt: now, UpdatedAt: now,
		},
	}

	responses := ToResponseList(announcements)

	require.Len(t, responses, 2)
	assert.Equal(t, "First", responses[0].Title)
	assert.Equal(t, "Second", responses[1].Title)
}

func TestToResponseList_Empty(t *testing.T) {
	responses := ToResponseList([]*entities.Announcement{})
	assert.Empty(t, responses)
}
