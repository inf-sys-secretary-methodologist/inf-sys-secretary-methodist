package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagFromEntity(t *testing.T) {
	now := time.Now()
	color := "#FF0000"
	tag := &entities.DocumentTag{
		ID:        1,
		Name:      "Urgent",
		Color:     &color,
		CreatedAt: now,
	}

	output := TagFromEntity(tag)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Urgent", output.Name)
	assert.Equal(t, &color, output.Color)
	assert.Equal(t, now, output.CreatedAt)
}

func TestTagFromEntity_NoColor(t *testing.T) {
	tag := &entities.DocumentTag{
		ID:        2,
		Name:      "Normal",
		CreatedAt: time.Now(),
	}

	output := TagFromEntity(tag)
	assert.Nil(t, output.Color)
}

func TestTagsFromEntities(t *testing.T) {
	tags := []*entities.DocumentTag{
		{ID: 1, Name: "A", CreatedAt: time.Now()},
		{ID: 2, Name: "B", CreatedAt: time.Now()},
	}

	outputs := TagsFromEntities(tags)

	require.Len(t, outputs, 2)
	assert.Equal(t, "A", outputs[0].Name)
	assert.Equal(t, "B", outputs[1].Name)
}

func TestTagsFromEntities_Empty(t *testing.T) {
	outputs := TagsFromEntities([]*entities.DocumentTag{})
	assert.Empty(t, outputs)
}
