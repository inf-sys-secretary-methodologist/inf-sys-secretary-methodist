package dto

import (
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToTemplateResponse(t *testing.T) {
	desc := "A template"
	content := "Dear {{name}}, ..."
	dt := &entities.DocumentType{
		ID:              1,
		Name:            "Report",
		Code:            "RPT",
		Description:     &desc,
		TemplateContent: &content,
		TemplateVariables: []entities.TemplateVariable{
			{Name: "name", Label: "Name", Type: "string", Required: true},
		},
	}

	resp := ToTemplateResponse(dt)

	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, "Report", resp.Name)
	assert.Equal(t, "RPT", resp.Code)
	assert.Equal(t, &desc, resp.Description)
	assert.Equal(t, &content, resp.TemplateContent)
	assert.True(t, resp.HasTemplate)
	require.Len(t, resp.TemplateVariables, 1)
	assert.Equal(t, "name", resp.TemplateVariables[0].Name)
}

func TestToTemplateResponse_NoTemplate(t *testing.T) {
	dt := &entities.DocumentType{
		ID:   2,
		Name: "Letter",
		Code: "LTR",
	}

	resp := ToTemplateResponse(dt)
	assert.False(t, resp.HasTemplate)
	assert.Nil(t, resp.TemplateContent)
}

func TestToTemplateResponse_EmptyTemplate(t *testing.T) {
	empty := ""
	dt := &entities.DocumentType{
		ID:              3,
		Name:            "Memo",
		Code:            "MEM",
		TemplateContent: &empty,
	}

	resp := ToTemplateResponse(dt)
	assert.False(t, resp.HasTemplate)
}

func TestToTemplateListResponse(t *testing.T) {
	content := "Hello {{name}}"
	empty := ""
	types := []entities.DocumentType{
		{ID: 1, Name: "A", Code: "A", TemplateContent: &content},
		{ID: 2, Name: "B", Code: "B", TemplateContent: nil},
		{ID: 3, Name: "C", Code: "C", TemplateContent: &empty},
	}

	resp := ToTemplateListResponse(types)

	require.NotNil(t, resp)
	assert.Equal(t, 1, resp.Total)
	require.Len(t, resp.Templates, 1)
	assert.Equal(t, "A", resp.Templates[0].Name)
}

func TestToTemplateListResponse_Empty(t *testing.T) {
	resp := ToTemplateListResponse([]entities.DocumentType{})
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Templates)
}
