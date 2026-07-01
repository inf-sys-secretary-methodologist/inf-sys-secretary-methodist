package usecases

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

func TestToCustomReportFilter(t *testing.T) {
	userID := int64(42)
	isPublic := true
	input := dto.CustomReportFilterInput{
		DataSource: "documents",
		IsPublic:   &isPublic,
		Search:     "test",
		Page:       2,
		PageSize:   25,
	}

	filter := ToCustomReportFilter(input, &userID)

	assert.Equal(t, 2, filter.Page)
	assert.Equal(t, 25, filter.PageSize)
	assert.Equal(t, "test", filter.Search)
	assert.Equal(t, &userID, filter.CreatedBy)
	require.NotNil(t, filter.DataSource)
	assert.Equal(t, entities.DataSourceDocuments, *filter.DataSource)
	assert.Equal(t, &isPublic, filter.IsPublic)
}

func TestToCustomReportFilter_Defaults(t *testing.T) {
	input := dto.CustomReportFilterInput{
		Page:     0,
		PageSize: 0,
	}

	filter := ToCustomReportFilter(input, nil)

	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 10, filter.PageSize)
	assert.Nil(t, filter.CreatedBy)
	assert.Nil(t, filter.DataSource)
}
