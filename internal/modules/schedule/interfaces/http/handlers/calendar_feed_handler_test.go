package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type fakeFeedSvc struct {
	tok       *entities.CalendarFeedToken
	getErr    error
	issueErr  error
	deleteErr error
	renderErr error
	ics       string
	rotated   bool
	deleted   bool
}

func (f *fakeFeedSvc) EnsureToken(_ context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	if f.issueErr != nil {
		return nil, f.issueErr
	}
	return &entities.CalendarFeedToken{UserID: userID, Token: "tok-ensure"}, nil
}

func (f *fakeFeedSvc) RotateToken(_ context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	f.rotated = true
	if f.issueErr != nil {
		return nil, f.issueErr
	}
	return &entities.CalendarFeedToken{UserID: userID, Token: "tok-rotate"}, nil
}

func (f *fakeFeedSvc) GetToken(_ context.Context, _ int64) (*entities.CalendarFeedToken, error) {
	return f.tok, f.getErr
}

func (f *fakeFeedSvc) DeleteToken(_ context.Context, _ int64) error {
	f.deleted = true
	return f.deleteErr
}

func (f *fakeFeedSvc) RenderFeed(_ context.Context, _ string) (string, error) {
	return f.ics, f.renderErr
}

func setupRouter(svc CalendarFeedService, withUser bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCalendarFeedHandler(svc, "https://host.example/")
	if withUser {
		r.Use(func(c *gin.Context) { c.Set("user_id", int64(42)); c.Next() })
	}
	r.GET("/sub", h.GetSubscription)
	r.POST("/sub", h.CreateSubscription)
	r.POST("/sub/rotate", h.RotateSubscription)
	r.DELETE("/sub", h.DeleteSubscription)
	r.GET("/public/calendar/:token/feed.ics", h.ServeFeed)
	return r
}

func do(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestServeFeed_OK(t *testing.T) {
	svc := &fakeFeedSvc{ics: "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"}
	w := do(setupRouter(svc, false), http.MethodGet, "/public/calendar/sometoken/feed.ics")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/calendar")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "schedule.ics")
	assert.Contains(t, w.Header().Get("Cache-Control"), "no-store")
	assert.Equal(t, svc.ics, w.Body.String())
}

func TestServeFeed_NotFound(t *testing.T) {
	svc := &fakeFeedSvc{renderErr: entities.ErrCalendarFeedTokenNotFound}
	w := do(setupRouter(svc, false), http.MethodGet, "/public/calendar/bad/feed.ics")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServeFeed_InternalError(t *testing.T) {
	svc := &fakeFeedSvc{renderErr: errors.New("boom")}
	w := do(setupRouter(svc, false), http.MethodGet, "/public/calendar/x/feed.ics")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSubscription_NotSubscribed(t *testing.T) {
	svc := &fakeFeedSvc{getErr: entities.ErrCalendarFeedTokenNotFound}
	w := do(setupRouter(svc, true), http.MethodGet, "/sub")

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Data struct {
			Subscribed bool   `json:"subscribed"`
			URL        string `json:"url"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.False(t, body.Data.Subscribed)
	assert.Empty(t, body.Data.URL)
}

func TestGetSubscription_Subscribed(t *testing.T) {
	svc := &fakeFeedSvc{tok: &entities.CalendarFeedToken{UserID: 42, Token: "secret123"}}
	w := do(setupRouter(svc, true), http.MethodGet, "/sub")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://host.example/api/public/calendar/secret123/feed.ics")
}

func TestGetSubscription_Unauthorized(t *testing.T) {
	svc := &fakeFeedSvc{}
	w := do(setupRouter(svc, false), http.MethodGet, "/sub")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateSubscription_ReturnsURL(t *testing.T) {
	svc := &fakeFeedSvc{}
	w := do(setupRouter(svc, true), http.MethodPost, "/sub")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "/api/public/calendar/tok-ensure/feed.ics")
}

func TestRotateSubscription_ReturnsNewURL(t *testing.T) {
	svc := &fakeFeedSvc{}
	w := do(setupRouter(svc, true), http.MethodPost, "/sub/rotate")

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.rotated)
	assert.Contains(t, w.Body.String(), "/api/public/calendar/tok-rotate/feed.ics")
}

func TestDeleteSubscription_NoContent(t *testing.T) {
	svc := &fakeFeedSvc{}
	w := do(setupRouter(svc, true), http.MethodDelete, "/sub")

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, svc.deleted)
}

func TestSubscription_ErrorPathsReturn500(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		svc    *fakeFeedSvc
	}{
		{"get error", http.MethodGet, "/sub", &fakeFeedSvc{getErr: errors.New("boom")}},
		{"create error", http.MethodPost, "/sub", &fakeFeedSvc{issueErr: errors.New("boom")}},
		{"rotate error", http.MethodPost, "/sub/rotate", &fakeFeedSvc{issueErr: errors.New("boom")}},
		{"delete error", http.MethodDelete, "/sub", &fakeFeedSvc{deleteErr: errors.New("boom")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := do(setupRouter(tc.svc, true), tc.method, tc.path)
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	}
}

func TestSubscription_UnauthenticatedOnAllMethods(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/sub"},
		{http.MethodPost, "/sub"},
		{http.MethodPost, "/sub/rotate"},
		{http.MethodDelete, "/sub"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := do(setupRouter(&fakeFeedSvc{}, false), tc.method, tc.path)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestFeedURL_NoTrailingSlashBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCalendarFeedHandler(&fakeFeedSvc{tok: &entities.CalendarFeedToken{Token: "abc"}}, "https://host.example") // no trailing slash
	r.Use(func(c *gin.Context) { c.Set("user_id", int64(42)); c.Next() })
	r.GET("/sub", h.GetSubscription)

	w := do(r, http.MethodGet, "/sub")
	assert.Contains(t, w.Body.String(), "https://host.example/api/public/calendar/abc/feed.ics")
}
