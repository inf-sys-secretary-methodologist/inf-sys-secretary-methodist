package usecases

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// SharingUseCase handles document sharing operations
type SharingUseCase struct {
	documentRepo   repositories.DocumentRepository
	permissionRepo repositories.PermissionRepository
	publicLinkRepo repositories.PublicLinkRepository
	auditLog       *logging.AuditLogger
	baseURL        string
}

// NewSharingUseCase creates a new SharingUseCase
func NewSharingUseCase(
	documentRepo repositories.DocumentRepository,
	permissionRepo repositories.PermissionRepository,
	publicLinkRepo repositories.PublicLinkRepository,
	auditLog *logging.AuditLogger,
	baseURL string,
) *SharingUseCase {
	return &SharingUseCase{
		documentRepo:   documentRepo,
		permissionRepo: permissionRepo,
		publicLinkRepo: publicLinkRepo,
		auditLog:       auditLog,
		baseURL:        baseURL,
	}
}

// ShareDocument shares a document with a user or role
func (uc *SharingUseCase) ShareDocument(ctx context.Context, input dto.ShareDocumentInput, grantedBy int64) (*dto.PermissionOutput, error) {
	// Check if document exists
	doc, err := uc.documentRepo.GetByID(ctx, input.DocumentID)
	if err != nil {
		return nil, err
	}

	// Check if granter has admin permission or is the author
	if doc.AuthorID != grantedBy {
		hasAdmin, err := uc.permissionRepo.HasPermission(ctx, input.DocumentID, grantedBy, entities.PermissionAdmin)
		if err != nil {
			return nil, err
		}
		if !hasAdmin {
			return nil, domainErrors.ErrForbidden
		}
	}

	// Check if permission already exists for this user/role
	if input.UserID != nil {
		existing, err := uc.permissionRepo.GetByDocumentAndUser(ctx, input.DocumentID, *input.UserID)
		if err == nil && existing != nil {
			// Update existing permission
			existing.Permission = entities.PermissionLevel(input.Permission)
			existing.ExpiresAt = input.ExpiresAt
			if err := uc.permissionRepo.Update(ctx, existing); err != nil {
				return nil, err
			}
			return dto.ToPermissionOutput(existing), nil
		}
	}

	// Create new permission
	var role *entities.UserRole
	if input.Role != nil {
		r := entities.UserRole(*input.Role)
		role = &r
	}

	permission := &entities.DocumentPermission{
		DocumentID: input.DocumentID,
		UserID:     input.UserID,
		Role:       role,
		Permission: entities.PermissionLevel(input.Permission),
		GrantedBy:  &grantedBy,
		ExpiresAt:  input.ExpiresAt,
	}

	if err := uc.permissionRepo.Create(ctx, permission); err != nil {
		return nil, err
	}

	// Log the action
	if uc.auditLog != nil {
		details := map[string]interface{}{
			"document_id": input.DocumentID,
			"permission":  input.Permission,
		}
		if input.UserID != nil {
			details["user_id"] = *input.UserID
		}
		if input.Role != nil {
			details["role"] = *input.Role
		}
		uc.auditLog.LogAuditEvent(ctx, "document_shared", "document_permission", details)
	}

	// Get full permission with user details
	return uc.GetPermission(ctx, permission.ID)
}

// RevokePermission revokes a permission
func (uc *SharingUseCase) RevokePermission(ctx context.Context, permissionID int64, revokedBy int64) error {
	// Get the permission
	permission, err := uc.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}

	// Check if revoker has admin permission or is the author
	doc, err := uc.documentRepo.GetByID(ctx, permission.DocumentID)
	if err != nil {
		return err
	}

	if doc.AuthorID != revokedBy {
		hasAdmin, err := uc.permissionRepo.HasPermission(ctx, permission.DocumentID, revokedBy, entities.PermissionAdmin)
		if err != nil {
			return err
		}
		if !hasAdmin {
			return domainErrors.ErrForbidden
		}
	}

	if err := uc.permissionRepo.Delete(ctx, permissionID); err != nil {
		return err
	}

	// Log the action
	if uc.auditLog != nil {
		uc.auditLog.LogAuditEvent(ctx, "permission_revoked", "document_permission", map[string]interface{}{
			"permission_id": permissionID,
			"document_id":   permission.DocumentID,
			"user_id":       revokedBy,
		})
	}

	return nil
}

// GetPermission retrieves a permission by ID
func (uc *SharingUseCase) GetPermission(ctx context.Context, id int64) (*dto.PermissionOutput, error) {
	permission, err := uc.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.ToPermissionOutput(permission), nil
}

// GetDocumentPermissions retrieves all permissions for a document
func (uc *SharingUseCase) GetDocumentPermissions(ctx context.Context, documentID int64, userID int64) ([]*dto.PermissionOutput, error) {
	// Check if user has access to view permissions
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	if doc.AuthorID != userID {
		hasPermission, err := uc.permissionRepo.HasAnyPermission(ctx, documentID, userID)
		if err != nil {
			return nil, err
		}
		if !hasPermission && !doc.IsPublic {
			return nil, domainErrors.ErrForbidden
		}
	}

	permissions, err := uc.permissionRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	return dto.ToPermissionOutputList(permissions), nil
}

// CheckUserPermission checks if a user has a specific permission level for a document
func (uc *SharingUseCase) CheckUserPermission(ctx context.Context, documentID, userID int64, permission entities.PermissionLevel) (bool, error) {
	// Check if user is author
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return false, err
	}

	if doc.AuthorID == userID {
		return true, nil // Author has full access
	}

	// Check if document is public and permission is read
	if doc.IsPublic && permission == entities.PermissionRead {
		return true, nil
	}

	return uc.permissionRepo.HasPermission(ctx, documentID, userID, permission)
}

// CreatePublicLink creates a public link for a document
func (uc *SharingUseCase) CreatePublicLink(ctx context.Context, input dto.CreatePublicLinkInput, createdBy int64) (*dto.PublicLinkOutput, error) {
	// Check if document exists
	doc, err := uc.documentRepo.GetByID(ctx, input.DocumentID)
	if err != nil {
		return nil, err
	}

	// Check if creator has permission
	if doc.AuthorID != createdBy {
		hasWrite, err := uc.permissionRepo.HasPermission(ctx, input.DocumentID, createdBy, entities.PermissionWrite)
		if err != nil {
			return nil, err
		}
		if !hasWrite {
			return nil, domainErrors.ErrForbidden
		}
	}

	// Generate token
	token, err := entities.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash password if provided
	var passwordHash *string
	if input.Password != nil && *input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		h := string(hash)
		passwordHash = &h
	}

	link := &entities.PublicLink{
		DocumentID:   input.DocumentID,
		Token:        token,
		Permission:   entities.PublicLinkPermission(input.Permission),
		CreatedBy:    createdBy,
		ExpiresAt:    input.ExpiresAt,
		MaxUses:      input.MaxUses,
		UseCount:     0,
		PasswordHash: passwordHash,
		IsActive:     true,
	}

	if err := uc.publicLinkRepo.Create(ctx, link); err != nil {
		return nil, err
	}

	// Log the action
	if uc.auditLog != nil {
		uc.auditLog.LogAuditEvent(ctx, "public_link_created", "public_link", map[string]interface{}{
			"document_id": input.DocumentID,
			"link_id":     link.ID,
			"permission":  input.Permission,
			"user_id":     createdBy,
		})
	}

	return dto.ToPublicLinkOutput(link, uc.baseURL), nil
}

// GetPublicLink retrieves a public link by ID
func (uc *SharingUseCase) GetPublicLink(ctx context.Context, id int64, userID int64) (*dto.PublicLinkOutput, error) {
	link, err := uc.publicLinkRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has access
	doc, err := uc.documentRepo.GetByID(ctx, link.DocumentID)
	if err != nil {
		return nil, err
	}

	if doc.AuthorID != userID && link.CreatedBy != userID {
		hasPermission, err := uc.permissionRepo.HasAnyPermission(ctx, link.DocumentID, userID)
		if err != nil {
			return nil, err
		}
		if !hasPermission {
			return nil, domainErrors.ErrForbidden
		}
	}

	return dto.ToPublicLinkOutput(link, uc.baseURL), nil
}

// GetDocumentPublicLinks retrieves all public links for a document
func (uc *SharingUseCase) GetDocumentPublicLinks(ctx context.Context, documentID int64, userID int64) ([]*dto.PublicLinkOutput, error) {
	// Check if user has access
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	if doc.AuthorID != userID {
		hasPermission, err := uc.permissionRepo.HasAnyPermission(ctx, documentID, userID)
		if err != nil {
			return nil, err
		}
		if !hasPermission {
			return nil, domainErrors.ErrForbidden
		}
	}

	links, err := uc.publicLinkRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	return dto.ToPublicLinkOutputList(links, uc.baseURL), nil
}

// DeactivatePublicLink deactivates a public link
func (uc *SharingUseCase) DeactivatePublicLink(ctx context.Context, linkID int64, userID int64) error {
	link, err := uc.publicLinkRepo.GetByID(ctx, linkID)
	if err != nil {
		return err
	}

	// Check if user has permission
	doc, err := uc.documentRepo.GetByID(ctx, link.DocumentID)
	if err != nil {
		return err
	}

	if doc.AuthorID != userID && link.CreatedBy != userID {
		hasAdmin, err := uc.permissionRepo.HasPermission(ctx, link.DocumentID, userID, entities.PermissionAdmin)
		if err != nil {
			return err
		}
		if !hasAdmin {
			return domainErrors.ErrForbidden
		}
	}

	if err := uc.publicLinkRepo.Deactivate(ctx, linkID); err != nil {
		return err
	}

	// Log the action
	if uc.auditLog != nil {
		uc.auditLog.LogAuditEvent(ctx, "public_link_deactivated", "public_link", map[string]interface{}{
			"link_id":     linkID,
			"document_id": link.DocumentID,
			"user_id":     userID,
		})
	}

	return nil
}

// DeletePublicLink deletes a public link
func (uc *SharingUseCase) DeletePublicLink(ctx context.Context, linkID int64, userID int64) error {
	link, err := uc.publicLinkRepo.GetByID(ctx, linkID)
	if err != nil {
		return err
	}

	// Check if user has permission
	doc, err := uc.documentRepo.GetByID(ctx, link.DocumentID)
	if err != nil {
		return err
	}

	if doc.AuthorID != userID && link.CreatedBy != userID {
		hasAdmin, err := uc.permissionRepo.HasPermission(ctx, link.DocumentID, userID, entities.PermissionAdmin)
		if err != nil {
			return err
		}
		if !hasAdmin {
			return domainErrors.ErrForbidden
		}
	}

	if err := uc.publicLinkRepo.Delete(ctx, linkID); err != nil {
		return err
	}

	// Log the action
	if uc.auditLog != nil {
		uc.auditLog.LogAuditEvent(ctx, "public_link_deleted", "public_link", map[string]interface{}{
			"link_id":     linkID,
			"document_id": link.DocumentID,
			"user_id":     userID,
		})
	}

	return nil
}

// AccessPublicLink provides access to a document via public link
func (uc *SharingUseCase) AccessPublicLink(ctx context.Context, token string, password *string) (*dto.DocumentAccessOutput, error) {
	link, err := uc.publicLinkRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check if link is valid
	if !link.IsValid() {
		return nil, domainErrors.ErrForbidden
	}

	// Check password if required
	if link.HasPassword() {
		if password == nil || *password == "" {
			return nil, domainErrors.NewDomainError("password_required", "Password required to access this link", nil)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*link.PasswordHash), []byte(*password)); err != nil {
			return nil, domainErrors.ErrUnauthorized
		}
	}

	// Get document
	doc, err := uc.documentRepo.GetByID(ctx, link.DocumentID)
	if err != nil {
		return nil, err
	}

	// Increment use count
	if err := uc.publicLinkRepo.IncrementUseCount(ctx, link.ID); err != nil {
		// Log but don't fail
		if uc.auditLog != nil {
			uc.auditLog.LogAuditEvent(ctx, "public_link_use_count_error", "public_link", map[string]interface{}{
				"link_id": link.ID,
				"error":   err.Error(),
			})
		}
	}

	return &dto.DocumentAccessOutput{
		ID:                 doc.ID,
		Title:              doc.Title,
		Subject:            doc.Subject,
		Content:            doc.Content,
		AuthorName:         doc.AuthorName,
		RegistrationNumber: doc.RegistrationNumber,
		RegistrationDate:   doc.RegistrationDate,
		FileName:           doc.FileName,
		FileSize:           doc.FileSize,
		MimeType:           doc.MimeType,
		CanDownload:        link.CanDownload(),
		CreatedAt:          doc.CreatedAt,
	}, nil
}

// GetSharedDocuments retrieves documents shared with a user
func (uc *SharingUseCase) GetSharedDocuments(ctx context.Context, filter dto.SharedDocumentsFilter) ([]*entities.Document, error) {
	// Get permissions by both user ID and role
	permissions, err := uc.permissionRepo.GetByUserIDOrRole(ctx, filter.UserID, filter.UserRole)
	if err != nil {
		return nil, err
	}

	var documents []*entities.Document
	for _, p := range permissions {
		if filter.Permission != nil && string(p.Permission) != *filter.Permission {
			continue
		}
		doc, err := uc.documentRepo.GetByID(ctx, p.DocumentID)
		if err != nil {
			continue // Skip deleted or inaccessible documents
		}
		documents = append(documents, doc)
	}

	// Apply pagination
	start := filter.Offset
	if start > len(documents) {
		return []*entities.Document{}, nil
	}
	end := start + filter.Limit
	if end > len(documents) || filter.Limit == 0 {
		end = len(documents)
	}

	return documents[start:end], nil
}

// GetMySharedDocuments retrieves documents that the user owns and has shared with others
func (uc *SharingUseCase) GetMySharedDocuments(ctx context.Context, userID int64, limit, offset int) ([]*dto.MySharedDocumentOutput, error) {
	// Get all permissions granted by this user (documents they own)
	permissions, err := uc.permissionRepo.GetByGrantedBy(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Group permissions by document
	docPermissions := make(map[int64][]*entities.DocumentPermission)
	for _, p := range permissions {
		docPermissions[p.DocumentID] = append(docPermissions[p.DocumentID], p)
	}

	// Build output
	result := make([]*dto.MySharedDocumentOutput, 0, len(docPermissions))
	for docID, perms := range docPermissions {
		doc, err := uc.documentRepo.GetByID(ctx, docID)
		if err != nil {
			continue // Skip deleted documents
		}

		// Only include documents owned by this user
		if doc.AuthorID != userID {
			continue
		}

		sharedWith := make([]dto.SharedWithInfo, 0, len(perms))
		for _, p := range perms {
			info := dto.SharedWithInfo{
				PermissionID: p.ID,
				Permission:   string(p.Permission),
				GrantedAt:    p.CreatedAt,
				ExpiresAt:    p.ExpiresAt,
			}
			if p.UserID != nil {
				info.UserID = p.UserID
				info.UserName = p.UserName
				info.UserEmail = p.UserEmail
			}
			if p.Role != nil {
				role := string(*p.Role)
				info.Role = &role
			}
			sharedWith = append(sharedWith, info)
		}

		result = append(result, &dto.MySharedDocumentOutput{
			DocumentID:    doc.ID,
			DocumentTitle: doc.Title,
			SharedWith:    sharedWith,
		})
	}

	// Apply pagination
	start := offset
	if start > len(result) {
		return []*dto.MySharedDocumentOutput{}, nil
	}
	end := start + limit
	if end > len(result) || limit == 0 {
		end = len(result)
	}

	return result[start:end], nil
}
