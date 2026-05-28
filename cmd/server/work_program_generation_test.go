package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	curEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

type fakeDisciplineReader struct {
	item  *curEntities.DisciplineItem
	err   error
	gotID int64
}

func (f *fakeDisciplineReader) GetByID(_ context.Context, id int64) (*curEntities.DisciplineItem, error) {
	f.gotID = id
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func TestControlFormLabel(t *testing.T) {
	cases := []struct {
		cf   curEntities.ControlForm
		want string
	}{
		{curEntities.ControlFormExam, "экзамен"},
		{curEntities.ControlFormZachet, "зачёт"},
		{curEntities.ControlFormDifferentialZachet, "дифференцированный зачёт"},
		{curEntities.ControlFormCourseProject, "курсовой проект"},
		{curEntities.ControlForm("seminar"), "seminar"}, // unknown → raw enum fallback
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, controlFormLabel(tc.cf), "label for %s", tc.cf)
	}
}

func sampleDisciplineItem() *curEntities.DisciplineItem {
	ts := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	return curEntities.ReconstituteDisciplineItem(
		7, 3, "Базы данных и СУБД",
		32, 48, 16, 24,
		curEntities.ControlFormExam,
		5, 4, 0, 1,
		ts, ts,
	)
}

func TestDisciplineInfoAdapter_MapsItem(t *testing.T) {
	reader := &fakeDisciplineReader{item: sampleDisciplineItem()}
	adapter := newDisciplineInfoAdapter(reader)

	info, err := adapter.GetDisciplineInfo(context.Background(), 7)
	require.NoError(t, err)

	assert.Equal(t, int64(7), reader.gotID)
	assert.Equal(t, "Базы данных и СУБД", info.Name)
	assert.Equal(t, 32, info.HoursLecture)
	assert.Equal(t, 48, info.HoursPractice)
	assert.Equal(t, 16, info.HoursLab)
	assert.Equal(t, 24, info.HoursSelfStudy)
	assert.Equal(t, "экзамен", info.ControlForm)
}

func TestDisciplineInfoAdapter_PropagatesError(t *testing.T) {
	sentinel := errors.New("discipline not found")
	adapter := newDisciplineInfoAdapter(&fakeDisciplineReader{err: sentinel})

	_, err := adapter.GetDisciplineInfo(context.Background(), 7)
	assert.ErrorIs(t, err, sentinel)
}
