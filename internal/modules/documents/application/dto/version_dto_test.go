package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToDocumentVersionOutput(t *testing.T) {
	now := time.Now()
	title := "Title"
	status := "approved"
	fileName := "v1.pdf"
	changedByName := "Admin"
	changeDesc := "Updated content"

	v := &entities.DocumentVersion{
		ID:                1,
		DocumentID:        10,
		Version:           2,
		Title:             &title,
		Status:            &status,
		FileName:          &fileName,
		ChangedBy:         42,
		ChangedByName:     &changedByName,
		ChangeDescription: &changeDesc,
		CreatedAt:         now,
	}

	output := ToDocumentVersionOutput(v)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(10), output.DocumentID)
	assert.Equal(t, 2, output.Version)
	assert.Equal(t, &title, output.Title)
	assert.Equal(t, &status, output.Status)
	assert.Equal(t, &fileName, output.FileName)
	assert.Equal(t, int64(42), output.ChangedBy)
	assert.Equal(t, &changedByName, output.ChangedByName)
	assert.Equal(t, &changeDesc, output.ChangeDescription)
}

func TestToDocumentVersionOutput_Nil(t *testing.T) {
	output := ToDocumentVersionOutput(nil)
	assert.Nil(t, output)
}

func TestToDocumentVersionOutputList(t *testing.T) {
	now := time.Now()
	versions := []*entities.DocumentVersion{
		{ID: 1, DocumentID: 10, Version: 1, ChangedBy: 1, CreatedAt: now},
		{ID: 2, DocumentID: 10, Version: 2, ChangedBy: 1, CreatedAt: now},
	}

	outputs := ToDocumentVersionOutputList(versions)

	require.Len(t, outputs, 2)
	assert.Equal(t, 1, outputs[0].Version)
	assert.Equal(t, 2, outputs[1].Version)
}

func TestToVersionDiffOutput(t *testing.T) {
	now := time.Now()
	d := &entities.DocumentVersionDiff{
		DocumentID:    10,
		FromVersion:   1,
		ToVersion:     2,
		ChangedFields: []string{"title", "content"},
		DiffData:      map[string]interface{}{"title": "changed"},
		CreatedAt:     now,
	}

	output := ToVersionDiffOutput(d)

	require.NotNil(t, output)
	assert.Equal(t, int64(10), output.DocumentID)
	assert.Equal(t, 1, output.FromVersion)
	assert.Equal(t, 2, output.ToVersion)
	assert.Equal(t, []string{"title", "content"}, output.ChangedFields)
	assert.Contains(t, output.DiffData, "title")
}

func TestToVersionDiffOutput_Nil(t *testing.T) {
	output := ToVersionDiffOutput(nil)
	assert.Nil(t, output)
}
