package http

// v0.153.11 Phase 6 #196 backfill — SyncHandler.Ping was at 0% because
// the existing handler test fixture passes nil odataClient (see
// handlers_test.go newSyncHandler). Pattern mirror к
// sync_usecase_test.go TestPing_Success / TestPing_Error: real
// httptest.Server backing real *odata.Client.

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/infrastructure/odata"
)

func newPingTestODataServer(t *testing.T) (*httptest.Server, *odata.Client) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/odata/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)

	cfg := &odata.Config{
		BaseURL:    server.URL + "/odata",
		Username:   "test",
		Password:   "test",
		Timeout:    2 * time.Second,
		MaxRetries: 0,
		RetryDelay: 0,
	}
	return server, odata.NewClient(cfg)
}

func newSyncHandlerWithOData(t *testing.T, client *odata.Client) *SyncHandler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	uc := usecases.NewSyncUseCase(
		client,
		new(mockSyncLogRepo),
		new(mockEmployeeRepo),
		new(mockStudentRepo),
		new(mockSyncConflictRepo),
		logger,
	)
	return NewSyncHandler(uc)
}

func TestSyncHandler_Ping_Success(t *testing.T) {
	server, client := newPingTestODataServer(t)
	defer server.Close()

	h := newSyncHandlerWithOData(t, client)
	router := setupRouter()
	router.GET("/sync/ping", h.Ping)

	w := performRequest(router, http.MethodGet, "/sync/ping", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	assert.Contains(t, w.Body.String(), "1C server is reachable")
}

func TestSyncHandler_Ping_Error(t *testing.T) {
	server, client := newPingTestODataServer(t)
	server.Close() // immediately close → Ping fails

	h := newSyncHandlerWithOData(t, client)
	router := setupRouter()
	router.GET("/sync/ping", h.Ping)

	w := performRequest(router, http.MethodGet, "/sync/ping", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "1C server is not reachable")
}
