package odata

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// Config holds the 1C OData client configuration
type Config struct {
	BaseURL          string        `json:"base_url"`          // e.g., http://server/config/odata/standard.odata
	Username         string        `json:"username"`          // 1C user
	Password         string        `json:"password"`          // 1C password
	Timeout          time.Duration `json:"timeout"`           // HTTP client timeout
	MaxRetries       int           `json:"max_retries"`       // Max retry attempts
	RetryDelay       time.Duration `json:"retry_delay"`       // Delay between retries
	EmployeesCatalog string        `json:"employees_catalog"` // Catalog name for employees
	StudentsCatalog  string        `json:"students_catalog"`  // Catalog name for students
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout:          30 * time.Second,
		MaxRetries:       3,
		RetryDelay:       1 * time.Second,
		EmployeesCatalog: "Catalog_Сотрудники",
		StudentsCatalog:  "Catalog_Студенты",
	}
}

// Client represents a 1C OData API client
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new 1C OData client
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// ODataResponse represents a generic OData response
type ODataResponse[T any] struct {
	Metadata string `json:"odata.metadata"`
	Value    []T    `json:"value"`
	NextLink string `json:"odata.nextLink,omitempty"`
}

// ODataError represents an OData error response
type ODataError struct {
	Error struct {
		Code    string `json:"code"`
		Message struct {
			Lang  string `json:"lang"`
			Value string `json:"value"`
		} `json:"message"`
	} `json:"odata.error"`
}

// basicAuth returns the Basic Auth header value
func (c *Client) basicAuth() string {
	auth := c.config.Username + ":" + c.config.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// buildURL builds the OData request URL
func (c *Client) buildURL(endpoint string, params map[string]string) string {
	baseURL := strings.TrimSuffix(c.config.BaseURL, "/")
	fullURL := baseURL + "/" + endpoint

	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		fullURL += "?" + values.Encode()
	}

	return fullURL
}

// doRequest performs an HTTP request with retries
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", c.basicAuth())
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Retry on server errors
		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// parseResponse parses the OData response
func parseResponse[T any](resp *http.Response) (*ODataResponse[T], error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var odataErr ODataError
		if json.Unmarshal(body, &odataErr) == nil && odataErr.Error.Message.Value != "" {
			return nil, fmt.Errorf("odata error: %s - %s", odataErr.Error.Code, odataErr.Error.Message.Value)
		}
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ODataResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetEmployees fetches employees from 1C
func (c *Client) GetEmployees(ctx context.Context, filter string, top, skip int) ([]entities.ODataEmployee, string, error) {
	params := map[string]string{
		"$format": "application/json",
	}

	if filter != "" {
		params["$filter"] = filter
	}
	if top > 0 {
		params["$top"] = fmt.Sprintf("%d", top)
	}
	if skip > 0 {
		params["$skip"] = fmt.Sprintf("%d", skip)
	}

	url := c.buildURL(c.config.EmployeesCatalog, params)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch employees: %w", err)
	}

	result, err := parseResponse[entities.ODataEmployee](resp)
	if err != nil {
		return nil, "", err
	}

	return result.Value, result.NextLink, nil
}

// GetAllEmployees fetches all employees with pagination
func (c *Client) GetAllEmployees(ctx context.Context) ([]entities.ODataEmployee, error) {
	var allEmployees []entities.ODataEmployee
	pageSize := 100
	skip := 0

	for {
		employees, nextLink, err := c.GetEmployees(ctx, "", pageSize, skip)
		if err != nil {
			return nil, err
		}

		allEmployees = append(allEmployees, employees...)

		if nextLink == "" || len(employees) < pageSize {
			break
		}

		skip += pageSize
	}

	return allEmployees, nil
}

// GetEmployeeByID fetches a single employee by Ref_Key
func (c *Client) GetEmployeeByID(ctx context.Context, refKey string) (*entities.ODataEmployee, error) {
	endpoint := fmt.Sprintf("%s(guid'%s')", c.config.EmployeesCatalog, refKey)
	params := map[string]string{
		"$format": "application/json",
	}

	url := c.buildURL(endpoint, params)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employee: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var employee entities.ODataEmployee
	if err := json.NewDecoder(resp.Body).Decode(&employee); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &employee, nil
}

// GetStudents fetches students from 1C
func (c *Client) GetStudents(ctx context.Context, filter string, top, skip int) ([]entities.ODataStudent, string, error) {
	params := map[string]string{
		"$format": "application/json",
	}

	if filter != "" {
		params["$filter"] = filter
	}
	if top > 0 {
		params["$top"] = fmt.Sprintf("%d", top)
	}
	if skip > 0 {
		params["$skip"] = fmt.Sprintf("%d", skip)
	}

	url := c.buildURL(c.config.StudentsCatalog, params)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch students: %w", err)
	}

	result, err := parseResponse[entities.ODataStudent](resp)
	if err != nil {
		return nil, "", err
	}

	return result.Value, result.NextLink, nil
}

// GetAllStudents fetches all students with pagination
func (c *Client) GetAllStudents(ctx context.Context) ([]entities.ODataStudent, error) {
	var allStudents []entities.ODataStudent
	pageSize := 100
	skip := 0

	for {
		students, nextLink, err := c.GetStudents(ctx, "", pageSize, skip)
		if err != nil {
			return nil, err
		}

		allStudents = append(allStudents, students...)

		if nextLink == "" || len(students) < pageSize {
			break
		}

		skip += pageSize
	}

	return allStudents, nil
}

// GetStudentByID fetches a single student by Ref_Key
func (c *Client) GetStudentByID(ctx context.Context, refKey string) (*entities.ODataStudent, error) {
	endpoint := fmt.Sprintf("%s(guid'%s')", c.config.StudentsCatalog, refKey)
	params := map[string]string{
		"$format": "application/json",
	}

	url := c.buildURL(endpoint, params)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch student: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var student entities.ODataStudent
	if err := json.NewDecoder(resp.Body).Decode(&student); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &student, nil
}

// GetMetadata fetches the OData metadata document
func (c *Client) GetMetadata(ctx context.Context) (string, error) {
	url := c.buildURL("$metadata", nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.basicAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// Ping checks if the 1C server is reachable
func (c *Client) Ping(ctx context.Context) error {
	url := c.buildURL("", map[string]string{"$format": "application/json"})

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("1C server is not reachable: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetActiveEmployeesFilter returns OData filter for active employees
func GetActiveEmployeesFilter() string {
	return "DeletionMark eq false"
}

// GetActiveStudentsFilter returns OData filter for active students
func GetActiveStudentsFilter() string {
	return "DeletionMark eq false"
}

// GetEmployeesByDepartmentFilter returns OData filter for employees by department
func GetEmployeesByDepartmentFilter(department string) string {
	return fmt.Sprintf("Подразделение eq '%s' and DeletionMark eq false", department)
}

// GetStudentsByGroupFilter returns OData filter for students by group
func GetStudentsByGroupFilter(group string) string {
	return fmt.Sprintf("Группа eq '%s' and DeletionMark eq false", group)
}

// GetStudentsByFacultyFilter returns OData filter for students by faculty
func GetStudentsByFacultyFilter(faculty string) string {
	return fmt.Sprintf("Факультет eq '%s' and DeletionMark eq false", faculty)
}
