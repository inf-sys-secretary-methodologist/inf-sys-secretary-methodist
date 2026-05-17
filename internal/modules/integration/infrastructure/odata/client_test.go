package odata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// v0.153.2 #196 coverage push: odata/client.go was at 0% covered (395 LOC,
// biggest single coverage gap remaining). Tests use httptest.Server to
// stub the 1C OData endpoint.

func testConfig(baseURL string) *Config {
	return &Config{
		BaseURL:          baseURL,
		Username:         "tester",
		Password:         "secret",
		Timeout:          5 * time.Second,
		MaxRetries:       1,
		RetryDelay:       1 * time.Millisecond,
		EmployeesCatalog: "Catalog_Сотрудники",
		StudentsCatalog:  "Catalog_Студенты",
	}
}

// --- DefaultConfig + NewClient + basicAuth + buildURL ---

func TestDefaultConfig_ReturnsExpectedValues(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 1*time.Second, cfg.RetryDelay)
	assert.Equal(t, "Catalog_Сотрудники", cfg.EmployeesCatalog)
	assert.Equal(t, "Catalog_Студенты", cfg.StudentsCatalog)
}

func TestNewClient_StoresConfigAndHTTPClient(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BaseURL = "http://example.com"
	c := NewClient(cfg)
	require.NotNil(t, c)
	require.NotNil(t, c.httpClient)
	assert.Equal(t, cfg, c.config)
	assert.Equal(t, cfg.Timeout, c.httpClient.Timeout)
}

func TestBasicAuth_EncodesCorrectly(t *testing.T) {
	c := NewClient(&Config{Username: "user1", Password: "pass1"})
	got := c.basicAuth()
	// "user1:pass1" base64 = "dXNlcjE6cGFzczE="
	assert.Equal(t, "Basic dXNlcjE6cGFzczE=", got)
}

func TestBuildURL_TrimsTrailingSlashAndEncodesParams(t *testing.T) {
	c := NewClient(&Config{BaseURL: "http://1c.example.com/api/"})
	url := c.buildURL("Catalog_X", map[string]string{"$format": "application/json"})
	assert.Contains(t, url, "http://1c.example.com/api/Catalog_X")
	assert.Contains(t, url, "%24format=application%2Fjson")
}

func TestBuildURL_NoParams(t *testing.T) {
	c := NewClient(&Config{BaseURL: "http://1c.example.com/api"})
	url := c.buildURL("Catalog_X", nil)
	assert.Equal(t, "http://1c.example.com/api/Catalog_X", url)
}

// --- doRequest happy path ---

func TestDoRequest_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Basic dGVzdGVyOnNlY3JldA==", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":[]}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	resp, err := c.doRequest(context.Background(), http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()
}

func TestDoRequest_RetryOn500ThenSucceed(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":[]}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	resp, err := c.doRequest(context.Background(), http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()
	assert.GreaterOrEqual(t, calls.Load(), int32(2))
}

func TestDoRequest_MaxRetriesExceededOn500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	resp, err := c.doRequest(context.Background(), http.MethodGet, server.URL+"/", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "max retries exceeded")
}

func TestDoRequest_NetworkError(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1")) // port 1 — connection-refused
	resp, err := c.doRequest(context.Background(), http.MethodGet, "http://127.0.0.1:1/", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestDoRequest_ContextCanceledDuringBackoff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // force retry
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.RetryDelay = 200 * time.Millisecond
	cfg.MaxRetries = 3
	c := NewClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	resp, err := c.doRequest(ctx, http.MethodGet, server.URL+"/", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.Error(t, err)
}

func TestDoRequest_InvalidURLToNewRequest(t *testing.T) {
	c := NewClient(testConfig("http://example.com"))
	resp, err := c.doRequest(context.Background(), "BAD METHOD", "ht%tp://broken", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.Error(t, err)
}

// --- parseResponse ---

func TestParseResponse_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"odata.metadata": "meta",
			"value":          []map[string]string{{"Ref_Key": "abc"}},
			"odata.nextLink": "next",
		})
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	resp, err := c.doRequest(context.Background(), http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	result, err := parseResponse[entities.ODataEmployee](resp)
	require.NoError(t, err)
	assert.Equal(t, "meta", result.Metadata)
	assert.Equal(t, "next", result.NextLink)
	require.Len(t, result.Value, 1)
}

func TestParseResponse_ODataErrorEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"odata.error":{"code":"E001","message":{"lang":"ru","value":"Invalid"}}}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	resp, err := c.doRequest(context.Background(), http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	_, err = parseResponse[entities.ODataEmployee](resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "odata error: E001")
	assert.Contains(t, err.Error(), "Invalid")
}

func TestParseResponse_NonOKWithoutErrorEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("plain error text"))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	resp, err := c.doRequest(context.Background(), http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	_, err = parseResponse[entities.ODataEmployee](resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed with status 403")
}

func TestParseResponse_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	resp, err := c.doRequest(context.Background(), http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	_, err = parseResponse[entities.ODataEmployee](resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

// --- GetEmployees + GetStudents ---

func TestGetEmployees_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":[{"Ref_Key":"e1","Code":"001"}],"odata.nextLink":""}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	emps, next, err := c.GetEmployees(context.Background(), "", 10, 0)
	require.NoError(t, err)
	require.Len(t, emps, 1)
	assert.Equal(t, "e1", emps[0].RefKey)
	assert.Empty(t, next)
}

func TestGetEmployees_FilterAndPaginationParamsAdded(t *testing.T) {
	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":[]}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	_, _, err := c.GetEmployees(context.Background(), "DeletionMark eq false", 50, 100)
	require.NoError(t, err)
	assert.Contains(t, capturedQuery, "%24filter=DeletionMark")
	assert.Contains(t, capturedQuery, "%24top=50")
	assert.Contains(t, capturedQuery, "%24skip=100")
}

func TestGetEmployees_NetworkError(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1"))
	_, _, err := c.GetEmployees(context.Background(), "", 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch employees")
}

func TestGetEmployees_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"odata.error":{"code":"X","message":{"value":"Bad"}}}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	_, _, err := c.GetEmployees(context.Background(), "", 0, 0)
	require.Error(t, err)
}

func TestGetAllEmployees_StopsOnEmptyNextLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":[{"Ref_Key":"e1"}]}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	emps, err := c.GetAllEmployees(context.Background())
	require.NoError(t, err)
	require.Len(t, emps, 1)
}

func TestGetAllEmployees_Paginates(t *testing.T) {
	var mu sync.Mutex
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		page++
		current := page
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		if current == 1 {
			// First page — return full pageSize (100) records + nextLink to force continuation.
			items := strings.Repeat(`{"Ref_Key":"e"},`, 100)
			items = strings.TrimSuffix(items, ",")
			_, _ = w.Write([]byte(`{"value":[` + items + `],"odata.nextLink":"next"}`))
			return
		}
		// Second page — fewer than pageSize → loop exits.
		_, _ = w.Write([]byte(`{"value":[{"Ref_Key":"e-last"}]}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	emps, err := c.GetAllEmployees(context.Background())
	require.NoError(t, err)
	assert.Len(t, emps, 101)
}

func TestGetAllEmployees_ErrorPropagates(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1"))
	emps, err := c.GetAllEmployees(context.Background())
	require.Error(t, err)
	assert.Nil(t, emps)
}

func TestGetEmployeeByID_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "guid'abc-123'")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"Ref_Key":"abc-123","Code":"001"}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	emp, err := c.GetEmployeeByID(context.Background(), "abc-123")
	require.NoError(t, err)
	require.NotNil(t, emp)
	assert.Equal(t, "abc-123", emp.RefKey)
}

func TestGetEmployeeByID_NotFoundReturnsNilNilError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	emp, err := c.GetEmployeeByID(context.Background(), "missing")
	require.NoError(t, err, "404 must collapse to (nil, nil)")
	assert.Nil(t, emp)
}

func TestGetEmployeeByID_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("not allowed"))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	emp, err := c.GetEmployeeByID(context.Background(), "blocked")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 403")
	assert.Nil(t, emp)
}

func TestGetEmployeeByID_NetworkError(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1"))
	_, err := c.GetEmployeeByID(context.Background(), "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch employee")
}

func TestGetEmployeeByID_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	emp, err := c.GetEmployeeByID(context.Background(), "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode")
	assert.Nil(t, emp)
}

// Students mirrors of Employees tests — narrower set since shape parallels.

func TestGetStudents_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":[{"Ref_Key":"s1","Code":"S001"}]}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	students, _, err := c.GetStudents(context.Background(), "DeletionMark eq false", 10, 5)
	require.NoError(t, err)
	require.Len(t, students, 1)
}

func TestGetStudents_NetworkError(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1"))
	_, _, err := c.GetStudents(context.Background(), "", 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch students")
}

func TestGetAllStudents_StopsOnEmptyNextLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":[{"Ref_Key":"s1"}]}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	students, err := c.GetAllStudents(context.Background())
	require.NoError(t, err)
	require.Len(t, students, 1)
}

func TestGetAllStudents_ErrorPropagates(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1"))
	_, err := c.GetAllStudents(context.Background())
	require.Error(t, err)
}

func TestGetStudentByID_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"Ref_Key":"sid","Code":"S001"}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	s, err := c.GetStudentByID(context.Background(), "sid")
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, "sid", s.RefKey)
}

func TestGetStudentByID_NotFoundReturnsNilNilError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	s, err := c.GetStudentByID(context.Background(), "missing")
	require.NoError(t, err)
	assert.Nil(t, s)
}

func TestGetStudentByID_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	_, err := c.GetStudentByID(context.Background(), "blocked")
	require.Error(t, err)
}

func TestGetStudentByID_NetworkError(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1"))
	_, err := c.GetStudentByID(context.Background(), "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch student")
}

func TestGetStudentByID_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	_, err := c.GetStudentByID(context.Background(), "id")
	require.Error(t, err)
}

// --- GetMetadata + Ping ---

func TestGetMetadata_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "$metadata")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<edmx:Edmx></edmx:Edmx>"))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	meta, err := c.GetMetadata(context.Background())
	require.NoError(t, err)
	assert.Contains(t, meta, "edmx:Edmx")
}

func TestGetMetadata_NetworkError(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1"))
	_, err := c.GetMetadata(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch metadata")
}

func TestPing_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(testConfig(server.URL))
	require.NoError(t, c.Ping(context.Background()))
}

func TestPing_NetworkError(t *testing.T) {
	c := NewClient(testConfig("http://127.0.0.1:1"))
	err := c.Ping(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not reachable")
}

// --- Filter helpers ---

func TestGetActiveEmployeesFilter(t *testing.T) {
	assert.Equal(t, "DeletionMark eq false", GetActiveEmployeesFilter())
}

func TestGetActiveStudentsFilter(t *testing.T) {
	assert.Equal(t, "DeletionMark eq false", GetActiveStudentsFilter())
}

func TestGetEmployeesByDepartmentFilter(t *testing.T) {
	f := GetEmployeesByDepartmentFilter("ИТ")
	assert.Contains(t, f, "Подразделение eq 'ИТ'")
	assert.Contains(t, f, "DeletionMark eq false")
}

func TestGetStudentsByGroupFilter(t *testing.T) {
	f := GetStudentsByGroupFilter("ИС-21")
	assert.Contains(t, f, "Группа eq 'ИС-21'")
}

func TestGetStudentsByFacultyFilter(t *testing.T) {
	f := GetStudentsByFacultyFilter("Информатика")
	assert.Contains(t, f, "Факультет eq 'Информатика'")
}
