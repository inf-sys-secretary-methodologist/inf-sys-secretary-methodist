package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPermissionOutput(t *testing.T) {
	now := time.Now()
	userID := int64(5)
	role := entities.RoleTeacher
	userName := "Teacher"
	userEmail := "teacher@example.com"
	grantedBy := int64(1)
	grantedByName := "Admin"
	expiresAt := now.Add(24 * time.Hour)

	perm := &entities.DocumentPermission{
		ID:            1,
		DocumentID:    10,
		UserID:        &userID,
		Role:          &role,
		Permission:    entities.PermissionWrite,
		GrantedBy:     &grantedBy,
		ExpiresAt:     &expiresAt,
		CreatedAt:     now,
		UserName:      &userName,
		UserEmail:     &userEmail,
		GrantedByName: &grantedByName,
	}

	output := ToPermissionOutput(perm)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(10), output.DocumentID)
	assert.Equal(t, &userID, output.UserID)
	roleStr := "teacher"
	assert.Equal(t, &roleStr, output.Role)
	assert.Equal(t, "write", output.Permission)
	assert.Equal(t, &grantedBy, output.GrantedBy)
	assert.Equal(t, &grantedByName, output.GrantedByName)
	assert.Equal(t, &expiresAt, output.ExpiresAt)
}

func TestToPermissionOutput_NilRole(t *testing.T) {
	perm := &entities.DocumentPermission{
		ID:         1,
		DocumentID: 10,
		Permission: entities.PermissionRead,
		CreatedAt:  time.Now(),
	}

	output := ToPermissionOutput(perm)
	assert.Nil(t, output.Role)
}

func TestToPermissionOutputList(t *testing.T) {
	now := time.Now()
	perms := []*entities.DocumentPermission{
		{ID: 1, DocumentID: 10, Permission: entities.PermissionRead, CreatedAt: now},
		{ID: 2, DocumentID: 10, Permission: entities.PermissionWrite, CreatedAt: now},
	}

	outputs := ToPermissionOutputList(perms)

	require.Len(t, outputs, 2)
	assert.Equal(t, "read", outputs[0].Permission)
	assert.Equal(t, "write", outputs[1].Permission)
}

func TestToPublicLinkOutput(t *testing.T) {
	now := time.Now()
	docTitle := "My Doc"
	createdByName := "Admin"
	maxUses := 10
	passHash := "somehash"

	link := &entities.PublicLink{
		ID:            1,
		DocumentID:    10,
		Token:         "abc123",
		Permission:    entities.PublicLinkDownload,
		CreatedBy:     42,
		ExpiresAt:     nil,
		MaxUses:       &maxUses,
		UseCount:      3,
		PasswordHash:  &passHash,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		DocumentTitle: &docTitle,
		CreatedByName: &createdByName,
	}

	output := ToPublicLinkOutput(link, "https://example.com")

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(10), output.DocumentID)
	assert.Equal(t, "abc123", output.Token)
	assert.Equal(t, "https://example.com/api/v1/public/documents/abc123", output.URL)
	assert.Equal(t, "download", output.Permission)
	assert.Equal(t, int64(42), output.CreatedBy)
	assert.Equal(t, &maxUses, output.MaxUses)
	assert.Equal(t, 3, output.UseCount)
	assert.True(t, output.HasPassword)
	assert.True(t, output.IsActive)
	assert.Equal(t, &docTitle, output.DocumentTitle)
}

func TestToPublicLinkOutput_NoPassword(t *testing.T) {
	link := &entities.PublicLink{
		ID:         1,
		DocumentID: 10,
		Token:      "xyz",
		Permission: entities.PublicLinkRead,
		CreatedBy:  1,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	output := ToPublicLinkOutput(link, "https://example.com")
	assert.False(t, output.HasPassword)
}

func TestToPublicLinkOutputList(t *testing.T) {
	now := time.Now()
	links := []*entities.PublicLink{
		{ID: 1, DocumentID: 10, Token: "a", Permission: entities.PublicLinkRead, CreatedBy: 1, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 2, DocumentID: 10, Token: "b", Permission: entities.PublicLinkDownload, CreatedBy: 1, IsActive: true, CreatedAt: now, UpdatedAt: now},
	}

	outputs := ToPublicLinkOutputList(links, "https://example.com")

	require.Len(t, outputs, 2)
	assert.Contains(t, outputs[0].URL, "/a")
	assert.Contains(t, outputs[1].URL, "/b")
}
