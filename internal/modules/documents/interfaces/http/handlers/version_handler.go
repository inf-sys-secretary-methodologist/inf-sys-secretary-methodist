// Package http contains HTTP request handlers for the documents module.
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// VersionHandler handles HTTP requests for document version endpoints.
type VersionHandler struct {
	usecase   *usecases.DocumentVersionUseCase
	validator *validation.Validator
}

// NewVersionHandler creates a new version handler.
func NewVersionHandler(usecase *usecases.DocumentVersionUseCase) *VersionHandler {
	return &VersionHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
	}
}

// GetVersions handles GET /api/documents/:id/versions
// @Summary Get all versions of a document
// @Description Retrieves version history for a specific document
// @Tags Document Versions
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.Response{data=dto.DocumentVersionListOutput}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/documents/{id}/versions [get]
func (h *VersionHandler) GetVersions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	output, err := h.usecase.GetVersions(c.Request.Context(), documentID, userID.(int64))
	if err != nil {
		resp := response.NotFound("Документ не найден")
		c.JSON(http.StatusNotFound, resp)
		return
	}

	resp := response.Success(output)
	c.JSON(http.StatusOK, resp)
}

// GetVersion handles GET /api/documents/:id/versions/:version
// @Summary Get a specific version of a document
// @Description Retrieves details of a specific version
// @Tags Document Versions
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param version path int true "Version number"
// @Success 200 {object} response.Response{data=dto.DocumentVersionOutput}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/documents/{id}/versions/{version} [get]
func (h *VersionHandler) GetVersion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		resp := response.BadRequest("Неверный номер версии")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	output, err := h.usecase.GetVersion(c.Request.Context(), documentID, version, userID.(int64))
	if err != nil {
		resp := response.NotFound("Версия не найдена")
		c.JSON(http.StatusNotFound, resp)
		return
	}

	resp := response.Success(output)
	c.JSON(http.StatusOK, resp)
}

// CreateVersion handles POST /api/documents/:id/versions
// @Summary Create a manual version snapshot
// @Description Creates a new version snapshot of the current document state
// @Tags Document Versions
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param input body dto.CreateVersionInput true "Version input"
// @Success 201 {object} response.Response{data=dto.DocumentVersionOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/documents/{id}/versions [post]
func (h *VersionHandler) CreateVersion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.CreateVersionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	description := "Ручная версия"
	if input.ChangeDescription != nil {
		description = *input.ChangeDescription
	}

	output, err := h.usecase.CreateVersion(c.Request.Context(), documentID, userID.(int64), description)
	if err != nil {
		resp := response.InternalError("Ошибка при создании версии: " + err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := response.Success(output)
	c.JSON(http.StatusCreated, resp)
}

// RestoreVersion handles POST /api/documents/:id/versions/:version/restore
// @Summary Restore document to a previous version
// @Description Restores the document to a specific previous version
// @Tags Document Versions
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param version path int true "Version number to restore"
// @Success 200 {object} response.Response{data=dto.DocumentOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/documents/{id}/versions/{version}/restore [post]
func (h *VersionHandler) RestoreVersion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		resp := response.BadRequest("Неверный номер версии")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	output, err := h.usecase.RestoreVersion(c.Request.Context(), documentID, version, userID.(int64))
	if err != nil {
		resp := response.InternalError("Ошибка при восстановлении версии: " + err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := response.Success(output)
	c.JSON(http.StatusOK, resp)
}

// CompareVersions handles GET /api/documents/:id/versions/compare
// @Summary Compare two versions
// @Description Compares two versions and returns the differences
// @Tags Document Versions
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param from query int true "From version number"
// @Param to query int true "To version number"
// @Success 200 {object} response.Response{data=dto.VersionDiffOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/documents/{id}/versions/compare [get]
func (h *VersionHandler) CompareVersions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	fromVersion, err := strconv.Atoi(c.Query("from"))
	if err != nil {
		resp := response.BadRequest("Неверный параметр 'from'")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	toVersion, err := strconv.Atoi(c.Query("to"))
	if err != nil {
		resp := response.BadRequest("Неверный параметр 'to'")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	output, err := h.usecase.CompareVersions(c.Request.Context(), documentID, fromVersion, toVersion, userID.(int64))
	if err != nil {
		resp := response.InternalError("Ошибка при сравнении версий: " + err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := response.Success(output)
	c.JSON(http.StatusOK, resp)
}

// DeleteVersion handles DELETE /api/documents/:id/versions/:version
// @Summary Delete a specific version
// @Description Deletes a specific version (cannot delete current version)
// @Tags Document Versions
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param version path int true "Version number to delete"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/documents/{id}/versions/{version} [delete]
func (h *VersionHandler) DeleteVersion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		resp := response.BadRequest("Неверный номер версии")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err = h.usecase.DeleteVersion(c.Request.Context(), documentID, version, userID.(int64))
	if err != nil {
		resp := response.InternalError("Ошибка при удалении версии: " + err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := response.Success(map[string]string{"message": "Версия успешно удалена"})
	c.JSON(http.StatusOK, resp)
}

// GetVersionFile handles GET /api/documents/:id/versions/:version/file
// @Summary Get file from a specific version
// @Description Gets file information and download URL from a specific version
// @Tags Document Versions
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param version path int true "Version number"
// @Success 200 {object} response.Response{data=dto.VersionFileDownloadOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/documents/{id}/versions/{version}/file [get]
func (h *VersionHandler) GetVersionFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		resp := response.BadRequest("Неверный номер версии")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	output, err := h.usecase.GetVersionFile(c.Request.Context(), documentID, version, userID.(int64))
	if err != nil {
		resp := response.NotFound("Файл версии не найден")
		c.JSON(http.StatusNotFound, resp)
		return
	}

	resp := response.Success(output)
	c.JSON(http.StatusOK, resp)
}
