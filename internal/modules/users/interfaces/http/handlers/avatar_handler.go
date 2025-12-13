// Package handlers contains HTTP request handlers for the users module.
package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

const (
	// MaxAvatarSize is the maximum allowed avatar file size (5MB)
	MaxAvatarSize = 5 * 1024 * 1024
	// AvatarFolder is the folder name in S3 bucket for avatars
	AvatarFolder = "avatars"
	// AvatarURLExpiration is how long presigned URLs for avatars are valid
	AvatarURLExpiration = 7 * 24 * time.Hour // 7 days
)

// AllowedAvatarTypes contains allowed MIME types for avatars
var AllowedAvatarTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// AvatarHandler handles avatar upload/delete operations.
type AvatarHandler struct {
	userUseCase *usecases.UserUseCase
	s3Client    *storage.S3Client
}

// NewAvatarHandler creates a new avatar handler.
func NewAvatarHandler(userUseCase *usecases.UserUseCase, s3Client *storage.S3Client) *AvatarHandler {
	return &AvatarHandler{
		userUseCase: userUseCase,
		s3Client:    s3Client,
	}
}

// Upload handles POST /api/users/:id/avatar - uploads user avatar.
func (h *AvatarHandler) Upload(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Check if user can update this profile (self or admin)
	currentUserID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	// For now, users can only update their own avatar
	// Admin check can be added later
	if currentUserID.(int64) != userID {
		userRole, _ := c.Get("user_role")
		if userRole != "system_admin" {
			resp := response.Forbidden("Нет прав для изменения аватара другого пользователя")
			c.JSON(http.StatusForbidden, resp)
			return
		}
	}

	// Get file from form
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		resp := response.BadRequest("Файл аватара не найден в запросе")
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > MaxAvatarSize {
		resp := response.BadRequest("Размер файла превышает 5MB")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !AllowedAvatarTypes[contentType] {
		resp := response.BadRequest("Недопустимый тип файла. Разрешены: JPG, PNG, GIF, WebP")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromMimeType(contentType)
	}
	filename := fmt.Sprintf("%s/%d_%s%s", AvatarFolder, userID, uuid.New().String()[:8], ext)

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		resp := response.InternalError("Ошибка чтения файла")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	ctx := c.Request.Context()

	// Delete old avatar if exists
	user, err := h.userUseCase.GetUser(ctx, userID)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	if user.Avatar != "" {
		// Extract object key from stored path and delete
		oldKey := user.Avatar
		if strings.HasPrefix(oldKey, AvatarFolder) {
			_ = h.s3Client.Delete(ctx, oldKey) // Ignore error for old file
		}
	}

	// Upload new avatar to S3
	reader := bytes.NewReader(content)
	_, err = h.s3Client.Upload(ctx, filename, reader, int64(len(content)), contentType)
	if err != nil {
		resp := response.InternalError("Ошибка загрузки файла")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Generate presigned URL for access
	avatarURL, err := h.s3Client.GetPresignedURL(ctx, filename, AvatarURLExpiration)
	if err != nil {
		// Try to delete uploaded file on error
		_ = h.s3Client.Delete(ctx, filename)
		resp := response.InternalError("Ошибка генерации URL")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Update user profile with the storage key (not the presigned URL)
	// We store the key and generate fresh URLs when needed
	input := &dto.UpdateUserProfileInput{
		Avatar: filename,
	}
	if err := h.userUseCase.UpdateUserProfile(ctx, userID, input); err != nil {
		// Try to delete uploaded file on error
		_ = h.s3Client.Delete(ctx, filename)
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"avatar_url": avatarURL,
		"avatar_key": filename,
		"message":    "Аватар успешно загружен",
	}))
}

// Delete handles DELETE /api/users/:id/avatar - deletes user avatar.
func (h *AvatarHandler) Delete(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Check if user can update this profile
	currentUserID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	if currentUserID.(int64) != userID {
		userRole, _ := c.Get("user_role")
		if userRole != "system_admin" {
			resp := response.Forbidden("Нет прав для удаления аватара другого пользователя")
			c.JSON(http.StatusForbidden, resp)
			return
		}
	}

	ctx := c.Request.Context()

	// Get current user to find avatar key
	user, err := h.userUseCase.GetUser(ctx, userID)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	if user.Avatar == "" {
		resp := response.BadRequest("Аватар не установлен")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Delete file from S3
	if strings.HasPrefix(user.Avatar, AvatarFolder) {
		if err := h.s3Client.Delete(ctx, user.Avatar); err != nil {
			// Log error but continue - we still want to clear the avatar field
			fmt.Printf("Warning: failed to delete avatar file from S3: %v\n", err)
		}
	}

	// Clear avatar in profile
	input := &dto.UpdateUserProfileInput{
		Avatar: "",
	}
	if err := h.userUseCase.UpdateUserProfile(ctx, userID, input); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "Аватар успешно удален",
	}))
}

// GetAvatarURL handles GET /api/users/:id/avatar - returns fresh presigned URL for avatar.
func (h *AvatarHandler) GetAvatarURL(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()

	user, err := h.userUseCase.GetUser(ctx, userID)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	if user.Avatar == "" {
		c.JSON(http.StatusOK, response.Success(gin.H{
			"avatar_url": "",
		}))
		return
	}

	// Generate fresh presigned URL
	avatarURL, err := h.s3Client.GetPresignedURL(ctx, user.Avatar, AvatarURLExpiration)
	if err != nil {
		resp := response.InternalError("Ошибка генерации URL")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"avatar_url": avatarURL,
	}))
}

// getExtensionFromMimeType returns file extension for common image types.
func getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
