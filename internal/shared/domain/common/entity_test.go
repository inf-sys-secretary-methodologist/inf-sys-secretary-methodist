package common

import (
	"testing"
	"time"
)

func TestEntity_Touch(t *testing.T) {
	entity := Entity{
		ID:        "test-id",
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	oldUpdatedAt := entity.UpdatedAt
	time.Sleep(10 * time.Millisecond)
	entity.Touch()

	if !entity.UpdatedAt.After(oldUpdatedAt) {
		t.Errorf("expected UpdatedAt to be updated after Touch()")
	}
}
