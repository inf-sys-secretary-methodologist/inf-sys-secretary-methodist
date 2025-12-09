// Package http contains HTTP request handlers for the documents module.
package http

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// DocumentHandler handles HTTP requests for document endpoints.
type DocumentHandler struct {
	usecase       *usecases.DocumentUseCase
	validator     *validation.Validator
	sanitizer     *sanitization.Sanitizer
	fileValidator *storage.FileValidator
}

// NewDocumentHandler creates a new document handler.
func NewDocumentHandler(usecase *usecases.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{
		usecase:       usecase,
		validator:     validation.NewValidator(),
		sanitizer:     sanitization.NewSanitizer(),
		fileValidator: storage.NewFileValidator(storage.DefaultFileValidatorConfig()),
	}
}

// Create handles document creation
func (h *DocumentHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var input dto.CreateDocumentInput

	// Check if this is a multipart form (with file) or JSON
	contentType := c.ContentType()
	if contentType == "application/json" {
		if err := c.ShouldBindJSON(&input); err != nil {
			resp := response.BadRequest("Неверный формат запроса")
			c.JSON(http.StatusBadRequest, resp)
			return
		}
	} else {
		// Parse multipart form
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
			resp := response.BadRequest("Ошибка при обработке формы")
			c.JSON(http.StatusBadRequest, resp)
			return
		}

		input.Title = c.PostForm("title")
		typeID, _ := strconv.ParseInt(c.PostForm("document_type_id"), 10, 64)
		input.DocumentTypeID = typeID

		if catID := c.PostForm("category_id"); catID != "" {
			id, _ := strconv.ParseInt(catID, 10, 64)
			input.CategoryID = &id
		}
		if subject := c.PostForm("subject"); subject != "" {
			input.Subject = &subject
		}
		if content := c.PostForm("content"); content != "" {
			input.Content = &content
		}
		if recipientID := c.PostForm("recipient_id"); recipientID != "" {
			id, _ := strconv.ParseInt(recipientID, 10, 64)
			input.RecipientID = &id
		}
		if importance := c.PostForm("importance"); importance != "" {
			input.Importance = &importance
		}
		if isPublic := c.PostForm("is_public"); isPublic == "true" {
			input.IsPublic = true
		}

		// Handle file upload with validation
		file, err := c.FormFile("file")
		if err == nil {
			// Validate file before accepting
			fileReader, err := file.Open()
			if err == nil {
				defer fileReader.Close()

				// Read file header for magic bytes validation
				headerBytes := make([]byte, 8)
				n, _ := fileReader.Read(headerBytes)
				headerBytes = headerBytes[:n]

				contentType := file.Header.Get("Content-Type")
				if contentType == "" {
					contentType = "application/octet-stream"
				}

				validationResult, _ := h.fileValidator.ValidateFile(
					file.Filename,
					file.Size,
					contentType,
					bytes.NewReader(headerBytes),
				)

				if !validationResult.Valid {
					resp := response.BadRequest(strings.Join(validationResult.Errors, "; "))
					c.JSON(http.StatusBadRequest, resp)
					return
				}
			}
			input.File = file
		}
	}

	// Sanitize inputs
	input.Title = h.sanitizer.SanitizeString(input.Title)
	if input.Subject != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Subject)
		input.Subject = &sanitized
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	doc, err := h.usecase.Create(ctx, input, userID.(int64))
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(doc)
	c.JSON(http.StatusCreated, resp)
}

// GetByID handles getting a document by ID
func (h *DocumentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	doc, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(doc)
	c.JSON(http.StatusOK, resp)
}

// Update handles document update
func (h *DocumentHandler) Update(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateDocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	if input.Title != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Title)
		input.Title = &sanitized
	}
	if input.Subject != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Subject)
		input.Subject = &sanitized
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	doc, err := h.usecase.Update(ctx, id, input, userID.(int64))
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(doc)
	c.JSON(http.StatusOK, resp)
}

// Delete handles document deletion
func (h *DocumentHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.Delete(ctx, id, userID.(int64)); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{"message": "Документ успешно удален"})
	c.JSON(http.StatusOK, resp)
}

// List handles listing documents with filters
func (h *DocumentHandler) List(c *gin.Context) {
	var filter dto.DocumentFilterInput
	if err := c.ShouldBindQuery(&filter); err != nil {
		resp := response.BadRequest("Неверные параметры фильтрации")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Set defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	ctx := c.Request.Context()
	result, err := h.usecase.List(ctx, filter)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.List(result.Documents, response.Pagination{
		Page:       result.Page,
		PerPage:    result.PageSize,
		Total:      int(result.Total),
		TotalPages: result.TotalPages,
	})
	c.JSON(http.StatusOK, resp)
}

// UploadFile handles file upload to an existing document
func (h *DocumentHandler) UploadFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp := response.BadRequest("Файл не найден в запросе")
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Read file header for magic bytes validation
	headerBytes := make([]byte, 8)
	n, _ := file.Read(headerBytes)
	headerBytes = headerBytes[:n]

	// Reset file reader
	if _, err := file.Seek(0, 0); err != nil {
		resp := response.BadRequest("Ошибка при чтении файла")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Validate file
	validationResult, err := h.fileValidator.ValidateFile(
		header.Filename,
		header.Size,
		contentType,
		bytes.NewReader(headerBytes),
	)
	if err != nil {
		resp := response.InternalError("Ошибка при валидации файла")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	if !validationResult.Valid {
		resp := response.BadRequest(strings.Join(validationResult.Errors, "; "))
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	doc, err := h.usecase.UploadFile(ctx, id, file, validationResult.SanitizedName, header.Size, contentType, userID.(int64))
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(doc)
	c.JSON(http.StatusOK, resp)
}

// DownloadFile handles file download from a document
// Query params:
//   - inline=true: display file in browser instead of downloading (for preview)
func (h *DocumentHandler) DownloadFile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	reader, fileInfo, err := h.usecase.DownloadFile(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}
	defer reader.Close()

	// Check if inline viewing is requested (for preview in browser)
	isInline := c.Query("inline") == "true"
	if isInline {
		c.Header("Content-Disposition", "inline; filename=\""+fileInfo.FileName+"\"")
		// Delete X-Frame-Options header set by security middleware to allow cross-origin iframe
		// In development, frontend (3000) and backend (8080) are on different ports
		c.Writer.Header().Del("X-Frame-Options")
		c.Writer.Header().Del("Content-Security-Policy")
		// Allow framing from any origin for preview functionality
		c.Header("Content-Security-Policy", "frame-ancestors *")
	} else {
		c.Header("Content-Disposition", "attachment; filename=\""+fileInfo.FileName+"\"")
	}
	c.Header("Content-Type", fileInfo.ContentType)
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size, 10))

	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		// Log error but can't send response as headers already sent
		return
	}
}

// DeleteFile handles file deletion from a document
func (h *DocumentHandler) DeleteFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.DeleteFile(ctx, id, userID.(int64)); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{"message": "Файл успешно удален"})
	c.JSON(http.StatusOK, resp)
}

// GetDocumentTypes handles getting all document types
func (h *DocumentHandler) GetDocumentTypes(c *gin.Context) {
	ctx := c.Request.Context()
	types, err := h.usecase.GetDocumentTypes(ctx)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(types)
	c.JSON(http.StatusOK, resp)
}

// GetCategories handles getting all document categories
func (h *DocumentHandler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.usecase.GetCategories(ctx)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(categories)
	c.JSON(http.StatusOK, resp)
}

// Search handles full-text search for documents
// @Summary Search documents
// @Description Perform full-text search on documents with ranking and highlighting
// @Tags documents
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param document_type_id query int false "Filter by document type ID"
// @Param category_id query int false "Filter by category ID"
// @Param author_id query int false "Filter by author ID"
// @Param status query string false "Filter by status"
// @Param importance query string false "Filter by importance"
// @Param from_date query string false "Filter by creation date from (YYYY-MM-DD)"
// @Param to_date query string false "Filter by creation date to (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Results per page" default(20)
// @Success 200 {object} dto.SearchOutput
// @Router /documents/search [get]
func (h *DocumentHandler) Search(c *gin.Context) {
	var input dto.SearchInput
	if err := c.ShouldBindQuery(&input); err != nil {
		resp := response.BadRequest("Неверные параметры поиска")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Validate search query
	if input.Query == "" {
		resp := response.BadRequest("Параметр поиска (q) обязателен")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize search query
	input.Query = h.sanitizer.SanitizeString(input.Query)

	// Set defaults
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	ctx := c.Request.Context()
	result, err := h.usecase.Search(ctx, input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}
