// Package usecases contains business logic for the documents module.
package usecases

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// DocumentUseCase handles document business logic
type DocumentUseCase struct {
	documentRepo     repositories.DocumentRepository
	documentTypeRepo repositories.DocumentTypeRepository
	categoryRepo     repositories.DocumentCategoryRepository
	s3Client         *storage.S3Client
	auditLog         *logging.AuditLogger
}

// NewDocumentUseCase creates a new document use case
func NewDocumentUseCase(
	documentRepo repositories.DocumentRepository,
	documentTypeRepo repositories.DocumentTypeRepository,
	categoryRepo repositories.DocumentCategoryRepository,
	s3Client *storage.S3Client,
	auditLog *logging.AuditLogger,
) *DocumentUseCase {
	return &DocumentUseCase{
		documentRepo:     documentRepo,
		documentTypeRepo: documentTypeRepo,
		categoryRepo:     categoryRepo,
		s3Client:         s3Client,
		auditLog:         auditLog,
	}
}

// Create creates a new document
func (uc *DocumentUseCase) Create(ctx context.Context, input dto.CreateDocumentInput, authorID int64) (*dto.DocumentOutput, error) {
	// Validate document type exists
	docType, err := uc.documentTypeRepo.GetByID(ctx, input.DocumentTypeID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Create document entity
	doc := entities.NewDocument(input.Title, input.DocumentTypeID, authorID)
	doc.CategoryID = input.CategoryID
	doc.Subject = input.Subject
	doc.Content = input.Content
	doc.RecipientID = input.RecipientID
	doc.Deadline = input.Deadline
	doc.IsPublic = input.IsPublic

	if input.Importance != nil {
		doc.Importance = entities.DocumentImportance(*input.Importance)
	}

	// Handle file upload if provided
	if input.File != nil {
		fileInfo, err := uc.uploadFileFromMultipart(ctx, doc, input.File)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file: %w", err)
		}
		doc.SetFile(input.File.Filename, fileInfo.Key, fileInfo.ContentType, fileInfo.Size)
	}

	// Save document
	if err := uc.documentRepo.Create(ctx, doc); err != nil {
		// Rollback file upload on error
		if doc.FilePath != nil {
			_ = uc.s3Client.Delete(ctx, *doc.FilePath)
		}
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Log history
	_ = uc.documentRepo.AddHistory(ctx, &entities.DocumentHistory{
		DocumentID: doc.ID,
		UserID:     &authorID,
		Action:     "created",
	})

	// Log audit event
	uc.logAudit(ctx, "document_created", "document", map[string]interface{}{
		"document_id":      doc.ID,
		"title":            doc.Title,
		"document_type_id": doc.DocumentTypeID,
		"author_id":        authorID,
		"has_file":         doc.HasFile(),
	})

	output := dto.ToDocumentOutput(doc)
	output.DocumentTypeName = docType.Name
	return output, nil
}

// GetByID retrieves a document by ID
func (uc *DocumentUseCase) GetByID(ctx context.Context, id int64) (*dto.DocumentOutput, error) {
	doc, err := uc.documentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	return dto.ToDocumentOutput(doc), nil
}

// Update updates an existing document
func (uc *DocumentUseCase) Update(ctx context.Context, id int64, input dto.UpdateDocumentInput, userID int64) (*dto.DocumentOutput, error) {
	doc, err := uc.documentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Update fields
	if input.Title != nil {
		doc.Title = *input.Title
	}
	if input.Subject != nil {
		doc.Subject = input.Subject
	}
	if input.Content != nil {
		doc.Content = input.Content
	}
	if input.CategoryID != nil {
		doc.CategoryID = input.CategoryID
	}
	if input.RecipientID != nil {
		doc.RecipientID = input.RecipientID
	}
	if input.Deadline != nil {
		doc.Deadline = input.Deadline
	}
	if input.Importance != nil {
		doc.Importance = entities.DocumentImportance(*input.Importance)
	}
	if input.IsPublic != nil {
		doc.IsPublic = *input.IsPublic
	}

	if err := uc.documentRepo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	// Log history
	_ = uc.documentRepo.AddHistory(ctx, &entities.DocumentHistory{
		DocumentID: doc.ID,
		UserID:     &userID,
		Action:     "updated",
	})

	// Log audit event
	uc.logAudit(ctx, "document_updated", "document", map[string]interface{}{
		"document_id": doc.ID,
		"user_id":     userID,
	})

	return dto.ToDocumentOutput(doc), nil
}

// Delete soft deletes a document
func (uc *DocumentUseCase) Delete(ctx context.Context, id int64, userID int64) error {
	doc, err := uc.documentRepo.GetByID(ctx, id)
	if err != nil {
		return errors.ErrNotFound
	}

	if err := uc.documentRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Log history
	_ = uc.documentRepo.AddHistory(ctx, &entities.DocumentHistory{
		DocumentID: doc.ID,
		UserID:     &userID,
		Action:     "deleted",
	})

	// Log audit event
	uc.logAudit(ctx, "document_deleted", "document", map[string]interface{}{
		"document_id": id,
		"user_id":     userID,
		"title":       doc.Title,
	})

	return nil
}

// List retrieves documents with filters
func (uc *DocumentUseCase) List(ctx context.Context, filter dto.DocumentFilterInput) (*dto.DocumentListOutput, error) {
	// Convert DTO filter to repository filter
	repoFilter := repositories.DocumentFilter{
		AuthorID:        filter.AuthorID,
		RecipientID:     filter.RecipientID,
		DocumentTypeID:  filter.DocumentTypeID,
		CategoryID:      filter.CategoryID,
		IsPublic:        filter.IsPublic,
		SearchQuery:     filter.Search,
		FromDate:        filter.FromDate,
		ToDate:          filter.ToDate,
		Limit:           filter.PageSize,
		Offset:          (filter.Page - 1) * filter.PageSize,
		CurrentUserID:   filter.CurrentUserID,
		CurrentUserRole: filter.CurrentUserRole,
	}

	if filter.Status != nil {
		status := entities.DocumentStatus(*filter.Status)
		repoFilter.Status = &status
	}
	if filter.Importance != nil {
		importance := entities.DocumentImportance(*filter.Importance)
		repoFilter.Importance = &importance
	}
	if filter.OrderBy != nil {
		repoFilter.OrderBy = *filter.OrderBy
	} else {
		repoFilter.OrderBy = "created_at DESC"
	}

	docs, total, err := uc.documentRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	outputs := make([]*dto.DocumentOutput, len(docs))
	for i, doc := range docs {
		outputs[i] = dto.ToDocumentOutput(doc)
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &dto.DocumentListOutput{
		Documents:  outputs,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UploadFile uploads a file to an existing document
func (uc *DocumentUseCase) UploadFile(ctx context.Context, documentID int64, file io.Reader, fileName string, fileSize int64, contentType string, userID int64) (*dto.DocumentOutput, error) {
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Delete old file if exists
	if doc.HasFile() {
		_ = uc.s3Client.Delete(ctx, *doc.FilePath)
	}

	// Generate storage key
	key := storage.GenerateKey(doc.ID, fileName)

	// Upload to S3
	fileInfo, err := uc.s3Client.Upload(ctx, key, file, fileSize, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Update document
	doc.SetFile(fileName, fileInfo.Key, fileInfo.ContentType, fileInfo.Size)
	doc.Version++

	// Auto-register document when file is uploaded (if still in draft)
	if doc.Status == entities.DocumentStatusDraft {
		doc.Status = entities.DocumentStatusRegistered
	}

	if err := uc.documentRepo.Update(ctx, doc); err != nil {
		// Rollback
		_ = uc.s3Client.Delete(ctx, key)
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	// Create version record
	_ = uc.documentRepo.CreateVersion(ctx, &entities.DocumentVersion{
		DocumentID:        doc.ID,
		Version:           doc.Version,
		FileName:          doc.FileName,
		FilePath:          doc.FilePath,
		FileSize:          doc.FileSize,
		ChangedBy:         userID,
		ChangeDescription: strPtr("File uploaded"),
	})

	// Log history
	_ = uc.documentRepo.AddHistory(ctx, &entities.DocumentHistory{
		DocumentID: doc.ID,
		UserID:     &userID,
		Action:     "file_uploaded",
	})

	// Log audit event
	uc.logAudit(ctx, "document_file_uploaded", "document", map[string]interface{}{
		"document_id": documentID,
		"user_id":     userID,
		"file_name":   fileName,
		"file_size":   fileSize,
		"version":     doc.Version,
	})

	return dto.ToDocumentOutput(doc), nil
}

// DownloadFile returns file stream for download
func (uc *DocumentUseCase) DownloadFile(ctx context.Context, documentID int64) (io.ReadCloser, *dto.FileDownloadOutput, error) {
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, nil, errors.ErrNotFound
	}

	if !doc.HasFile() {
		return nil, nil, fmt.Errorf("document has no file attached: %w", errors.ErrNotFound)
	}

	reader, fileInfo, err := uc.s3Client.Download(ctx, *doc.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file: %w", err)
	}

	return reader, &dto.FileDownloadOutput{
		FileName:    *doc.FileName,
		ContentType: fileInfo.ContentType,
		Size:        fileInfo.Size,
	}, nil
}

// DeleteFile removes file from document
func (uc *DocumentUseCase) DeleteFile(ctx context.Context, documentID int64, userID int64) error {
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return errors.ErrNotFound
	}

	if !doc.HasFile() {
		return fmt.Errorf("document has no file attached: %w", errors.ErrNotFound)
	}

	// Delete from S3
	if err := uc.s3Client.Delete(ctx, *doc.FilePath); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// Clear file info
	doc.ClearFile()
	if err := uc.documentRepo.Update(ctx, doc); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	// Log history
	_ = uc.documentRepo.AddHistory(ctx, &entities.DocumentHistory{
		DocumentID: doc.ID,
		UserID:     &userID,
		Action:     "file_deleted",
	})

	// Log audit event
	uc.logAudit(ctx, "document_file_deleted", "document", map[string]interface{}{
		"document_id": documentID,
		"user_id":     userID,
	})

	return nil
}

// GetDocumentTypes returns all document types
func (uc *DocumentUseCase) GetDocumentTypes(ctx context.Context) ([]*dto.DocumentTypeOutput, error) {
	types, err := uc.documentTypeRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get document types: %w", err)
	}

	outputs := make([]*dto.DocumentTypeOutput, len(types))
	for i, t := range types {
		outputs[i] = &dto.DocumentTypeOutput{
			ID:                   t.ID,
			Name:                 t.Name,
			Code:                 t.Code,
			Description:          t.Description,
			RequiresApproval:     t.RequiresApproval,
			RequiresRegistration: t.RequiresRegistration,
		}
	}
	return outputs, nil
}

// GetCategories returns all document categories
func (uc *DocumentUseCase) GetCategories(ctx context.Context) ([]*dto.DocumentCategoryOutput, error) {
	categories, err := uc.categoryRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	outputs := make([]*dto.DocumentCategoryOutput, len(categories))
	for i, c := range categories {
		outputs[i] = &dto.DocumentCategoryOutput{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			ParentID:    c.ParentID,
		}
	}
	return outputs, nil
}

// Search performs full-text search on documents with ranking and highlighting
func (uc *DocumentUseCase) Search(ctx context.Context, input dto.SearchInput) (*dto.SearchOutput, error) {
	// Validate input
	if input.Query == "" {
		return nil, fmt.Errorf("search query is required: %w", errors.ErrValidationFailed)
	}

	// Set default pagination
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	// Convert DTO to repository filter
	filter := repositories.SearchFilter{
		Query:           input.Query,
		DocumentTypeID:  input.DocumentTypeID,
		CategoryID:      input.CategoryID,
		AuthorID:        input.AuthorID,
		FromDate:        input.FromDate,
		ToDate:          input.ToDate,
		IncludeDeleted:  false,
		Limit:           input.PageSize,
		Offset:          (input.Page - 1) * input.PageSize,
		CurrentUserID:   input.CurrentUserID,
		CurrentUserRole: input.CurrentUserRole,
	}

	if input.Status != nil {
		status := entities.DocumentStatus(*input.Status)
		filter.Status = &status
	}
	if input.Importance != nil {
		importance := entities.DocumentImportance(*input.Importance)
		filter.Importance = &importance
	}

	// Perform search
	results, total, err := uc.documentRepo.Search(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	// Convert to output DTOs
	outputs := make([]*dto.SearchResultOutput, len(results))
	for i, result := range results {
		outputs[i] = &dto.SearchResultOutput{
			Document:           dto.ToDocumentOutput(result.Document),
			Rank:               result.Rank,
			HighlightedTitle:   result.HighlightedTitle,
			HighlightedSubject: result.HighlightedSubject,
			HighlightedContent: result.HighlightedContent,
		}
	}

	// Calculate total pages
	totalPages := int(total) / input.PageSize
	if int(total)%input.PageSize > 0 {
		totalPages++
	}

	// Log audit event
	uc.logAudit(ctx, "document_search", "document", map[string]interface{}{
		"query":         input.Query,
		"results_count": total,
	})

	return &dto.SearchOutput{
		Results:    outputs,
		Query:      input.Query,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Helper methods

func (uc *DocumentUseCase) uploadFileFromMultipart(ctx context.Context, doc *entities.Document, file *multipart.FileHeader) (*storage.FileInfo, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	key := storage.GenerateKey(doc.ID, file.Filename)
	contentType := detectContentType(file.Filename)

	return uc.s3Client.Upload(ctx, key, reader, file.Size, contentType)
}

func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentTypes := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
	}
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func strPtr(s string) *string {
	return &s
}

// logAudit safely logs an audit event with nil check
func (uc *DocumentUseCase) logAudit(ctx context.Context, action, resourceType string, details map[string]interface{}) {
	if uc.auditLog != nil {
		uc.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}
