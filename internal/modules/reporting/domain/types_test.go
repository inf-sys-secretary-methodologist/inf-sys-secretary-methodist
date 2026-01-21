package domain

import "testing"

func TestReportStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status ReportStatus
		want   bool
	}{
		{"draft", ReportStatusDraft, true},
		{"generating", ReportStatusGenerating, true},
		{"ready", ReportStatusReady, true},
		{"reviewing", ReportStatusReviewing, true},
		{"approved", ReportStatusApproved, true},
		{"rejected", ReportStatusRejected, true},
		{"published", ReportStatusPublished, true},
		{"invalid", ReportStatus("invalid"), false},
		{"empty", ReportStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputFormat_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
		want   bool
	}{
		{"pdf", OutputFormatPDF, true},
		{"xlsx", OutputFormatXLSX, true},
		{"docx", OutputFormatDOCX, true},
		{"html", OutputFormatHTML, true},
		{"invalid", OutputFormat("invalid"), false},
		{"empty", OutputFormat(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeriodType_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		periodType PeriodType
		want       bool
	}{
		{"daily", PeriodTypeDaily, true},
		{"weekly", PeriodTypeWeekly, true},
		{"monthly", PeriodTypeMonthly, true},
		{"quarterly", PeriodTypeQuarterly, true},
		{"annual", PeriodTypeAnnual, true},
		{"invalid", PeriodType("invalid"), false},
		{"empty", PeriodType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.periodType.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReportStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   ReportStatus
		expected string
	}{
		{"draft", ReportStatusDraft, "draft"},
		{"generating", ReportStatusGenerating, "generating"},
		{"ready", ReportStatusReady, "ready"},
		{"reviewing", ReportStatusReviewing, "reviewing"},
		{"approved", ReportStatusApproved, "approved"},
		{"rejected", ReportStatusRejected, "rejected"},
		{"published", ReportStatusPublished, "published"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.status)
			}
		})
	}
}

func TestOutputFormatConstants(t *testing.T) {
	tests := []struct {
		name     string
		format   OutputFormat
		expected string
	}{
		{"pdf", OutputFormatPDF, "pdf"},
		{"xlsx", OutputFormatXLSX, "xlsx"},
		{"docx", OutputFormatDOCX, "docx"},
		{"html", OutputFormatHTML, "html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.format) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.format)
			}
		})
	}
}
