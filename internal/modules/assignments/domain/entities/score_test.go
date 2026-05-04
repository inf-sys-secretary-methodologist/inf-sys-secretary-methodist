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
		max       int
		wantErr   error
		wantValue int
		wantMax   int
	}{
		{name: "zero is valid lower boundary", value: 0, max: 100, wantValue: 0, wantMax: 100},
		{name: "value equals max is valid upper boundary", value: 100, max: 100, wantValue: 100, wantMax: 100},
		{name: "mid-range value is valid", value: 50, max: 100, wantValue: 50, wantMax: 100},
		{name: "small max accepts value at max", value: 5, max: 5, wantValue: 5, wantMax: 5},

		{name: "negative value is invalid", value: -1, max: 100, wantErr: entities.ErrInvalidScore},
		{name: "value above max is invalid", value: 101, max: 100, wantErr: entities.ErrInvalidScore},
		{name: "max of zero is invalid", value: 0, max: 0, wantErr: entities.ErrInvalidScore},
		{name: "negative max is invalid", value: 0, max: -5, wantErr: entities.ErrInvalidScore},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := entities.NewScore(tc.value, tc.max)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected error wrapping %v, got %v", tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantValue, got.Value())
			assert.Equal(t, tc.wantMax, got.Max())
		})
	}
}
