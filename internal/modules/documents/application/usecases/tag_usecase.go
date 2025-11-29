// Package usecases contains business logic for the documents module.
package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// TagUseCase handles business logic for document tags
type TagUseCase struct {
	tagRepo  repositories.DocumentTagRepository
	docRepo  repositories.DocumentRepository
	auditLog *logging.AuditLogger
}

// NewTagUseCase creates a new tag use case
func NewTagUseCase(tagRepo repositories.DocumentTagRepository, docRepo repositories.DocumentRepository, auditLog *logging.AuditLogger) *TagUseCase {
	return &TagUseCase{
		tagRepo:  tagRepo,
		docRepo:  docRepo,
		auditLog: auditLog,
	}
}

// Create creates a new tag
func (uc *TagUseCase) Create(ctx context.Context, input dto.CreateTagInput) (*dto.TagOutput, error) {
	// Check if tag with same name exists
	existing, _ := uc.tagRepo.GetByName(ctx, input.Name)
	if existing != nil {
		return nil, fmt.Errorf("тег с таким именем уже существует")
	}

	tag := &entities.DocumentTag{
		Name:  input.Name,
		Color: input.Color,
	}

	if err := uc.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}

	// Log audit event
	uc.logAudit(ctx, "tag_created", "tag", map[string]interface{}{
		"tag_id": tag.ID,
		"name":   tag.Name,
	})

	output := dto.TagFromEntity(tag)
	output.UsageCount = 0

	return output, nil
}

// Update updates an existing tag
func (uc *TagUseCase) Update(ctx context.Context, id int64, input dto.UpdateTagInput) (*dto.TagOutput, error) {
	tag, err := uc.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("тег не найден")
	}

	if input.Name != nil {
		// Check if another tag with same name exists
		existing, _ := uc.tagRepo.GetByName(ctx, *input.Name)
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("тег с таким именем уже существует")
		}
		tag.Name = *input.Name
	}

	if input.Color != nil {
		tag.Color = input.Color
	}

	if err := uc.tagRepo.Update(ctx, tag); err != nil {
		return nil, err
	}

	// Log audit event
	uc.logAudit(ctx, "tag_updated", "tag", map[string]interface{}{
		"tag_id": tag.ID,
		"name":   tag.Name,
	})

	output := dto.TagFromEntity(tag)
	usageCount, _ := uc.tagRepo.GetTagUsageCount(ctx, id)
	output.UsageCount = usageCount

	return output, nil
}

// Delete deletes a tag
func (uc *TagUseCase) Delete(ctx context.Context, id int64) error {
	tag, err := uc.tagRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("тег не найден")
	}

	if err := uc.tagRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Log audit event
	uc.logAudit(ctx, "tag_deleted", "tag", map[string]interface{}{
		"tag_id": id,
		"name":   tag.Name,
	})

	return nil
}

// GetByID retrieves a tag by ID
func (uc *TagUseCase) GetByID(ctx context.Context, id int64) (*dto.TagOutput, error) {
	tag, err := uc.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("тег не найден")
	}

	output := dto.TagFromEntity(tag)
	usageCount, _ := uc.tagRepo.GetTagUsageCount(ctx, id)
	output.UsageCount = usageCount

	return output, nil
}

// GetAll retrieves all tags
func (uc *TagUseCase) GetAll(ctx context.Context) ([]*dto.TagOutput, error) {
	tags, err := uc.tagRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]*dto.TagOutput, 0, len(tags))
	for _, t := range tags {
		output := dto.TagFromEntity(t)
		usageCount, _ := uc.tagRepo.GetTagUsageCount(ctx, t.ID)
		output.UsageCount = usageCount
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// Search searches for tags by name
func (uc *TagUseCase) Search(ctx context.Context, query string, limit int) ([]*dto.TagOutput, error) {
	tags, err := uc.tagRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	outputs := make([]*dto.TagOutput, 0, len(tags))
	for _, t := range tags {
		output := dto.TagFromEntity(t)
		usageCount, _ := uc.tagRepo.GetTagUsageCount(ctx, t.ID)
		output.UsageCount = usageCount
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// AddTagToDocument adds a tag to a document
func (uc *TagUseCase) AddTagToDocument(ctx context.Context, documentID, tagID int64) error {
	// Verify document exists
	doc, err := uc.docRepo.GetByID(ctx, documentID)
	if err != nil || doc == nil {
		return fmt.Errorf("документ не найден")
	}

	// Verify tag exists
	tag, err := uc.tagRepo.GetByID(ctx, tagID)
	if err != nil || tag == nil {
		return fmt.Errorf("тег не найден")
	}

	if err := uc.tagRepo.AddTagToDocument(ctx, documentID, tagID); err != nil {
		return err
	}

	// Log audit event
	uc.logAudit(ctx, "tag_added_to_document", "document_tag", map[string]interface{}{
		"document_id": documentID,
		"tag_id":      tagID,
	})

	return nil
}

// RemoveTagFromDocument removes a tag from a document
func (uc *TagUseCase) RemoveTagFromDocument(ctx context.Context, documentID, tagID int64) error {
	if err := uc.tagRepo.RemoveTagFromDocument(ctx, documentID, tagID); err != nil {
		return err
	}

	// Log audit event
	uc.logAudit(ctx, "tag_removed_from_document", "document_tag", map[string]interface{}{
		"document_id": documentID,
		"tag_id":      tagID,
	})

	return nil
}

// GetDocumentTags retrieves all tags for a document
func (uc *TagUseCase) GetDocumentTags(ctx context.Context, documentID int64) (*dto.DocumentTagsOutput, error) {
	// Verify document exists
	doc, err := uc.docRepo.GetByID(ctx, documentID)
	if err != nil || doc == nil {
		return nil, fmt.Errorf("документ не найден")
	}

	tags, err := uc.tagRepo.GetTagsByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	tagOutputs := make([]*dto.TagOutput, 0, len(tags))
	for _, t := range tags {
		output := dto.TagFromEntity(t)
		usageCount, _ := uc.tagRepo.GetTagUsageCount(ctx, t.ID)
		output.UsageCount = usageCount
		tagOutputs = append(tagOutputs, output)
	}

	return &dto.DocumentTagsOutput{
		DocumentID: documentID,
		Tags:       tagOutputs,
	}, nil
}

// SetDocumentTags replaces all tags for a document
func (uc *TagUseCase) SetDocumentTags(ctx context.Context, documentID int64, tagIDs []int64) (*dto.DocumentTagsOutput, error) {
	// Verify document exists
	doc, err := uc.docRepo.GetByID(ctx, documentID)
	if err != nil || doc == nil {
		return nil, fmt.Errorf("документ не найден")
	}

	// Verify all tags exist
	for _, tagID := range tagIDs {
		tag, err := uc.tagRepo.GetByID(ctx, tagID)
		if err != nil || tag == nil {
			return nil, fmt.Errorf("тег с ID %d не найден", tagID)
		}
	}

	if err := uc.tagRepo.SetDocumentTags(ctx, documentID, tagIDs); err != nil {
		return nil, err
	}

	// Log audit event
	uc.logAudit(ctx, "document_tags_set", "document_tag", map[string]interface{}{
		"document_id": documentID,
		"tag_ids":     tagIDs,
		"tag_count":   len(tagIDs),
	})

	return uc.GetDocumentTags(ctx, documentID)
}

// GetDocumentsByTag retrieves documents with a specific tag
func (uc *TagUseCase) GetDocumentsByTag(ctx context.Context, tagID int64, page, pageSize int) (*dto.TagWithDocumentsOutput, error) {
	tag, err := uc.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return nil, fmt.Errorf("тег не найден")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	documentIDs, total, err := uc.tagRepo.GetDocumentsByTagID(ctx, tagID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	tagOutput := dto.TagFromEntity(tag)
	tagOutput.UsageCount = total

	return &dto.TagWithDocumentsOutput{
		Tag:         tagOutput,
		DocumentIDs: documentIDs,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// logAudit safely logs an audit event with nil check
func (uc *TagUseCase) logAudit(ctx context.Context, action, resourceType string, details map[string]interface{}) {
	if uc.auditLog != nil {
		uc.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}
