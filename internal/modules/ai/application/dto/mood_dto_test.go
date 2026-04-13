package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToMoodResponse(t *testing.T) {
	now := time.Now()
	mood := &entities.MoodContext{
		State:            entities.MoodHappy,
		Intensity:        0.8,
		Reason:           "all good",
		OverdueDocuments: 2,
		AtRiskStudents:   1,
		ComputedAt:       now,
	}

	resp := ToMoodResponse(mood, "Great day!", "Hello!")

	require.NotNil(t, resp)
	assert.Equal(t, "happy", resp.State)
	assert.Equal(t, 0.8, resp.Intensity)
	assert.Equal(t, "all good", resp.Reason)
	assert.Equal(t, "Great day!", resp.Message)
	assert.Equal(t, "Hello!", resp.Greeting)
	assert.Equal(t, 2, resp.OverdueDocuments)
	assert.Equal(t, 1, resp.AtRiskStudents)
	assert.Equal(t, now, resp.ComputedAt)
	assert.Nil(t, resp.FunFact)
}
