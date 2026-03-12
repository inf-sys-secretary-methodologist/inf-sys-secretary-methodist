// Package handlers содержит HTTP обработчики модуля files.
package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// FileHandler обрабатывает HTTP запросы для управления файлами.
type FileHandler struct {
	fileUseCase    *usecases.FileUseCase
	versionUseCase *usecases.VersionUseCase
	validator      *validation.Validator
}

// NewFileHandler создаёт новый обработчик файлов.
func NewFileHandler(
	fileUseCase *usecases.FileUseCase,
	versionUseCase *usecases.VersionUseCase,
) *FileHandler {
	return &FileHandler{
		fileUseCase:    fileUseCase,
		versionUseCase: versionUseCase,
		validator:      validation.NewValidator(),
	}
}

// Upload обрабатывает POST /api/files/upload - загружает файл.
func (h *FileHandler) Upload(c *gin.Context) {
	// Получаем файл из формы
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp := response.BadRequest("Файл не найден в запросе")
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	defer func() { _ = file.Close() }()

	// Получаем user_id из контекста (должен быть установлен middleware авторизации)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	userID, _ := userIDVal.(int64)

	input := &dto.UploadFileInput{
		OriginalName: header.Filename,
		MimeType:     header.Header.Get("Content-Type"),
		Size:         header.Size,
		UserID:       userID,
	}

	// Если MIME не определён, используем default
	if input.MimeType == "" {
		input.MimeType = "application/octet-stream"
	}

	ctx := c.Request.Context()
	result, err := h.fileUseCase.UploadFile(ctx, file, input)
	if err != nil {
		var validErr *usecases.ValidationError
		if errors.As(err, &validErr) {
			resp := response.BadRequest(validErr.Message)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusCreated, response.Success(result))
}

// GetByID обрабатывает GET /api/files/:id - получает информацию о файле.
func (h *FileHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID файла")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.fileUseCase.GetFile(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// Download обрабатывает GET /api/files/:id/download - возвращает URL для скачивания.
func (h *FileHandler) Download(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID файла")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.fileUseCase.DownloadFile(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// Attach обрабатывает POST /api/files/:id/attach - прикрепляет файл к сущности.
func (h *FileHandler) Attach(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID файла")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.AttachFileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	input.FileID = id

	ctx := c.Request.Context()
	if err := h.fileUseCase.AttachFile(ctx, &input); err != nil {
		var validErr *usecases.ValidationError
		if errors.As(err, &validErr) {
			resp := response.BadRequest(validErr.Message)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Файл прикреплён"}))
}

// Delete обрабатывает DELETE /api/files/:id - удаляет файл.
func (h *FileHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID файла")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Получаем user_id из контекста
	userIDVal, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	userID, _ := userIDVal.(int64)

	ctx := c.Request.Context()
	if err := h.fileUseCase.DeleteFile(ctx, id, userID); err != nil {
		var permErr *usecases.PermissionError
		if errors.As(err, &permErr) {
			resp := response.Forbidden(permErr.Message)
			c.JSON(http.StatusForbidden, resp)
			return
		}
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Файл удалён"}))
}

// List обрабатывает GET /api/files - возвращает список файлов с пагинацией.
func (h *FileHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	ctx := c.Request.Context()
	result, err := h.fileUseCase.ListFiles(ctx, page, limit)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetByDocument обрабатывает GET /api/documents/:document_id/files - файлы документа.
func (h *FileHandler) GetByDocument(c *gin.Context) {
	documentID, err := strconv.ParseInt(c.Param("document_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.fileUseCase.GetFilesByDocument(ctx, documentID)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"files": result}))
}

// GetByTask обрабатывает GET /api/tasks/:task_id/files - файлы задачи.
func (h *FileHandler) GetByTask(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("task_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID задачи")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.fileUseCase.GetFilesByTask(ctx, taskID)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"files": result}))
}

// GetByAnnouncement обрабатывает GET /api/announcements/:announcement_id/files - файлы объявления.
func (h *FileHandler) GetByAnnouncement(c *gin.Context) {
	announcementID, err := strconv.ParseInt(c.Param("announcement_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID объявления")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.fileUseCase.GetFilesByAnnouncement(ctx, announcementID)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"files": result}))
}

// CreateVersion обрабатывает POST /api/files/:id/versions - создаёт новую версию.
func (h *FileHandler) CreateVersion(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID файла")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Получаем файл из формы
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp := response.BadRequest("Файл не найден в запросе")
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	defer func() { _ = file.Close() }()

	// Получаем user_id из контекста
	userIDVal, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	userID, _ := userIDVal.(int64)

	comment := c.PostForm("comment")

	input := &dto.CreateVersionInput{
		FileID:  id,
		Comment: comment,
		UserID:  userID,
	}

	ctx := c.Request.Context()
	result, err := h.versionUseCase.CreateVersion(ctx, file, header.Size, input)
	if err != nil {
		var validErr *usecases.ValidationError
		if errors.As(err, &validErr) {
			resp := response.BadRequest(validErr.Message)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusCreated, response.Success(result))
}

// GetVersions обрабатывает GET /api/files/:id/versions - получает все версии файла.
func (h *FileHandler) GetVersions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID файла")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.versionUseCase.GetVersions(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"versions": result}))
}

// DownloadVersion обрабатывает GET /api/files/:id/versions/:version - скачивает версию.
func (h *FileHandler) DownloadVersion(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID файла")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	versionNumber, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		resp := response.BadRequest("Неверный номер версии")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.versionUseCase.DownloadVersion(ctx, id, versionNumber)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// CleanupExpired обрабатывает POST /api/files/cleanup - очищает истёкшие временные файлы.
// Доступ должен быть ограничен администраторам или cron job.
func (h *FileHandler) CleanupExpired(c *gin.Context) {
	ctx := c.Request.Context()
	count, err := h.fileUseCase.CleanupExpiredFiles(ctx)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message":       "Очистка завершена",
		"deleted_count": count,
		"timestamp":     time.Now(),
	}))
}
