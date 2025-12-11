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

// SharingHandler handles document sharing HTTP requests
type SharingHandler struct {
	usecase   *usecases.SharingUseCase
	validator *validation.Validator
}

// NewSharingHandler creates a new SharingHandler
func NewSharingHandler(usecase *usecases.SharingUseCase, validator *validation.Validator) *SharingHandler {
	return &SharingHandler{
		usecase:   usecase,
		validator: validator,
	}
}

// ShareDocument shares a document with a user or role
// @Summary Share document
// @Description Share a document with a user or role
// @Tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param input body dto.ShareDocumentInput true "Share input"
// @Success 201 {object} response.Response{data=dto.PermissionOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /documents/{id}/share [post]
func (h *SharingHandler) ShareDocument(c *gin.Context) {
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

	var input dto.ShareDocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	input.DocumentID = documentID

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	permission, err := h.usecase.ShareDocument(ctx, input, userID.(int64))
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(permission)
	c.JSON(http.StatusCreated, resp)
}

// RevokePermission revokes a document permission
// @Summary Revoke permission
// @Description Revoke a document permission
// @Tags sharing
// @Produce json
// @Param id path int true "Document ID"
// @Param permissionId path int true "Permission ID"
// @Success 204
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /documents/{id}/permissions/{permissionId} [delete]
func (h *SharingHandler) RevokePermission(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	permissionID, err := strconv.ParseInt(c.Param("permissionId"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID права доступа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.RevokePermission(ctx, permissionID, userID.(int64)); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetDocumentPermissions retrieves all permissions for a document
// @Summary Get document permissions
// @Description Get all permissions for a document
// @Tags sharing
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.Response{data=[]dto.PermissionOutput}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /documents/{id}/permissions [get]
func (h *SharingHandler) GetDocumentPermissions(c *gin.Context) {
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

	ctx := c.Request.Context()
	permissions, err := h.usecase.GetDocumentPermissions(ctx, documentID, userID.(int64))
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(permissions)
	c.JSON(http.StatusOK, resp)
}

// CreatePublicLink creates a public link for a document
// @Summary Create public link
// @Description Create a public link for a document
// @Tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param input body dto.CreatePublicLinkInput true "Public link input"
// @Success 201 {object} response.Response{data=dto.PublicLinkOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /documents/{id}/public-links [post]
func (h *SharingHandler) CreatePublicLink(c *gin.Context) {
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

	var input dto.CreatePublicLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	input.DocumentID = documentID

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	link, err := h.usecase.CreatePublicLink(ctx, input, userID.(int64))
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(link)
	c.JSON(http.StatusCreated, resp)
}

// GetDocumentPublicLinks retrieves all public links for a document
// @Summary Get document public links
// @Description Get all public links for a document
// @Tags sharing
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.Response{data=[]dto.PublicLinkOutput}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /documents/{id}/public-links [get]
func (h *SharingHandler) GetDocumentPublicLinks(c *gin.Context) {
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

	ctx := c.Request.Context()
	links, err := h.usecase.GetDocumentPublicLinks(ctx, documentID, userID.(int64))
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(links)
	c.JSON(http.StatusOK, resp)
}

// DeactivatePublicLink deactivates a public link
// @Summary Deactivate public link
// @Description Deactivate a public link
// @Tags sharing
// @Produce json
// @Param id path int true "Document ID"
// @Param linkId path int true "Link ID"
// @Success 204
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /documents/{id}/public-links/{linkId}/deactivate [post]
func (h *SharingHandler) DeactivatePublicLink(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	linkID, err := strconv.ParseInt(c.Param("linkId"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID ссылки")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.DeactivatePublicLink(ctx, linkID, userID.(int64)); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.Status(http.StatusNoContent)
}

// DeletePublicLink deletes a public link
// @Summary Delete public link
// @Description Delete a public link
// @Tags sharing
// @Produce json
// @Param id path int true "Document ID"
// @Param linkId path int true "Link ID"
// @Success 204
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /documents/{id}/public-links/{linkId} [delete]
func (h *SharingHandler) DeletePublicLink(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	linkID, err := strconv.ParseInt(c.Param("linkId"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID ссылки")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.DeletePublicLink(ctx, linkID, userID.(int64)); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.Status(http.StatusNoContent)
}

// AccessPublicDocument provides access to a document via public link
// @Summary Access public document
// @Description Access a document via public link
// @Tags public
// @Accept json
// @Produce json
// @Param token path string true "Public link token"
// @Param input body dto.AccessPublicLinkInput false "Password if required"
// @Success 200 {object} response.Response{data=dto.DocumentAccessOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /public/documents/{token} [post]
func (h *SharingHandler) AccessPublicDocument(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		resp := response.BadRequest("Токен не указан")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.AccessPublicLinkInput
	// Password is optional, so don't fail if body is empty
	_ = c.ShouldBindJSON(&input)
	input.Token = token

	ctx := c.Request.Context()
	document, err := h.usecase.AccessPublicLink(ctx, token, input.Password)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(document)
	c.JSON(http.StatusOK, resp)
}

// GetSharedDocuments retrieves documents shared with the current user
// @Summary Get shared documents
// @Description Get documents shared with the current user
// @Tags sharing
// @Produce json
// @Param permission query string false "Filter by permission level"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} response.Response{data=[]dto.DocumentOutput}
// @Failure 401 {object} response.Response
// @Router /documents/shared [get]
func (h *SharingHandler) GetSharedDocuments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Get user role for role-based permission lookup
	var userRole string
	if role, exists := c.Get("role"); exists {
		userRole = role.(string)
	}

	filter := dto.SharedDocumentsFilter{
		UserID:   userID.(int64),
		UserRole: userRole,
		Limit:    20,
		Offset:   0,
	}

	if permission := c.Query("permission"); permission != "" {
		filter.Permission = &permission
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(c.Query("offset")); err == nil && offset >= 0 {
		filter.Offset = offset
	}

	ctx := c.Request.Context()
	documents, err := h.usecase.GetSharedDocuments(ctx, filter)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	// Convert to DTO for proper JSON serialization (includes has_file field)
	output := make([]*dto.DocumentOutput, len(documents))
	for i, doc := range documents {
		output[i] = dto.ToDocumentOutput(doc)
	}

	resp := response.Success(output)
	c.JSON(http.StatusOK, resp)
}

// GetMySharedDocuments retrieves documents that the current user has shared with others
// @Summary Get my shared documents
// @Description Get documents that the current user owns and has shared with others
// @Tags sharing
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} response.Response{data=[]dto.MySharedDocumentOutput}
// @Failure 401 {object} response.Response
// @Router /documents/my-shared [get]
func (h *SharingHandler) GetMySharedDocuments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	limit := 20
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	documents, err := h.usecase.GetMySharedDocuments(ctx, userID.(int64), limit, offset)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(documents)
	c.JSON(http.StatusOK, resp)
}
