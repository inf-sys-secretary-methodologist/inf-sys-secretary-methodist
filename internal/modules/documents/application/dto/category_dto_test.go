package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryFromEntity(t *testing.T) {
	now := time.Now()
	desc := "A category"
	parentID := int64(5)
	cat := &entities.DocumentCategory{
		ID:          1,
		Name:        "Reports",
		Description: &desc,
		ParentID:    &parentID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	output := CategoryFromEntity(cat)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Reports", output.Name)
	assert.Equal(t, &desc, output.Description)
	assert.Equal(t, &parentID, output.ParentID)
	assert.Equal(t, now, output.CreatedAt)
}

func TestCategoryTreeFromEntity(t *testing.T) {
	now := time.Now()
	child := &entities.CategoryTreeNode{
		ID:            2,
		Name:          "Child",
		DocumentCount: 3,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	parentID := int64(1)
	child.ParentID = &parentID

	node := &entities.CategoryTreeNode{
		ID:            1,
		Name:          "Parent",
		DocumentCount: 10,
		Children:      []*entities.CategoryTreeNode{child},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	output := CategoryTreeFromEntity(node)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Parent", output.Name)
	assert.Equal(t, int64(10), output.DocumentCount)
	require.Len(t, output.Children, 1)
	assert.Equal(t, "Child", output.Children[0].Name)
	assert.Equal(t, int64(3), output.Children[0].DocumentCount)
}

func TestCategoryTreeFromEntity_NoChildren(t *testing.T) {
	node := &entities.CategoryTreeNode{
		ID:        1,
		Name:      "Leaf",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	output := CategoryTreeFromEntity(node)
	assert.Nil(t, output.Children)
}
