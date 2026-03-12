package api

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

// ReportType represents a report type definition.
type ReportType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// Report represents a report resource.
type Report struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	ReportTypeID   int64  `json:"report_type_id"`
	ReportTypeName string `json:"report_type_name"`
	AuthorID       int64  `json:"author_id"`
	AuthorName     string `json:"author_name"`
	PeriodStart    string `json:"period_start"`
	PeriodEnd      string `json:"period_end"`
}

// ReportList represents a paginated list of reports.
type ReportList struct {
	Reports []Report `json:"reports"`
	Total   int      `json:"total"`
}

// GetReportTypes retrieves available report types.
func (c *Client) GetReportTypes(ctx context.Context, a *agent.Agent) ([]ReportType, error) {
	resp, err := c.Get(ctx, "/api/report-types", a)
	if err != nil {
		return nil, err
	}
	var types []ReportType
	if err := ParseData(resp, &types); err != nil {
		return nil, err
	}
	return types, nil
}

// CreateReportRequest represents a request to create a new report.
type CreateReportRequest struct {
	ReportTypeID int64  `json:"report_type_id"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
	IsPublic     bool   `json:"is_public,omitempty"`
}

// CreateReport creates a new report.
func (c *Client) CreateReport(ctx context.Context, a *agent.Agent, req CreateReportRequest) (*Report, error) {
	resp, err := c.Post(ctx, "/api/reports", a, req)
	if err != nil {
		return nil, fmt.Errorf("create report: %w", err)
	}
	var report Report
	if err := ParseData(resp, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReports retrieves reports.
func (c *Client) ListReports(ctx context.Context, a *agent.Agent, queryParams string) (*ReportList, error) {
	path := "/api/reports"
	if queryParams != "" {
		path += "?" + queryParams
	}
	resp, err := c.Get(ctx, path, a)
	if err != nil {
		return nil, err
	}
	var list ReportList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GenerateReport triggers report data generation.
func (c *Client) GenerateReport(ctx context.Context, a *agent.Agent, reportID int64) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/reports/%d/generate", reportID), a, nil)
	return err
}

// SubmitReport submits a report for review.
func (c *Client) SubmitReport(ctx context.Context, a *agent.Agent, reportID int64) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/reports/%d/submit", reportID), a, nil)
	return err
}

// ReviewReport reviews a report (approve or reject).
func (c *Client) ReviewReport(ctx context.Context, a *agent.Agent, reportID int64, action, comment string) error {
	body := map[string]any{
		"action":  action,
		"comment": comment,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/reports/%d/review", reportID), a, body)
	return err
}

// PublishReport publishes a report.
func (c *Client) PublishReport(ctx context.Context, a *agent.Agent, reportID int64, isPublic bool) error {
	body := map[string]any{
		"is_public": isPublic,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/reports/%d/publish", reportID), a, body)
	return err
}

// AddReportComment adds a comment to a report.
func (c *Client) AddReportComment(ctx context.Context, a *agent.Agent, reportID int64, content string) error {
	body := map[string]any{
		"content": content,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/reports/%d/comments", reportID), a, body)
	return err
}
