package entities_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

func TestNewScore_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		wantErr   error
		wantValue int
	}{
		{name: "zero is valid lower boundary", value: 0, wantValue: 0},
		{name: "positive value is valid", value: 50, wantValue: 50},
		{name: "large value is valid (upper bound is enforced by Assignment)", value: 100000, wantValue: 100000},

		{name: "negative value is invalid", value: -1, wantErr: entities.ErrInvalidScore},
		{name: "very negative value is invalid", value: -1000, wantErr: entities.ErrInvalidScore},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := entities.NewScore(tc.value)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected error wrapping %v, got %v", tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantValue, got.Value())
		})
	}
}
