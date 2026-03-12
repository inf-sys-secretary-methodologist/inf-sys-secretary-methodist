package domain

// ReportStatus represents the status of a report
type ReportStatus string

// ReportStatus values.
const (
	ReportStatusDraft      ReportStatus = "draft"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusReady      ReportStatus = "ready"
	ReportStatusReviewing  ReportStatus = "reviewing"
	ReportStatusApproved   ReportStatus = "approved"
	ReportStatusRejected   ReportStatus = "rejected"
	ReportStatusPublished  ReportStatus = "published"
)

// IsValid checks if the status is valid
func (s ReportStatus) IsValid() bool {
	switch s {
	case ReportStatusDraft, ReportStatusGenerating, ReportStatusReady,
		ReportStatusReviewing, ReportStatusApproved, ReportStatusRejected, ReportStatusPublished:
		return true
	}
	return false
}

// OutputFormat represents the output format of a report
type OutputFormat string

// OutputFormat values.
const (
	OutputFormatPDF  OutputFormat = "pdf"
	OutputFormatXLSX OutputFormat = "xlsx"
	OutputFormatDOCX OutputFormat = "docx"
	OutputFormatHTML OutputFormat = "html"
)

// IsValid checks if the format is valid
func (f OutputFormat) IsValid() bool {
	switch f {
	case OutputFormatPDF, OutputFormatXLSX, OutputFormatDOCX, OutputFormatHTML:
		return true
	}
	return false
}

// PeriodType represents the period type for periodic reports
type PeriodType string

// PeriodType values.
const (
	PeriodTypeDaily     PeriodType = "daily"
	PeriodTypeWeekly    PeriodType = "weekly"
	PeriodTypeMonthly   PeriodType = "monthly"
	PeriodTypeQuarterly PeriodType = "quarterly"
	PeriodTypeAnnual    PeriodType = "annual"
)

// IsValid checks if the period type is valid
func (p PeriodType) IsValid() bool {
	switch p {
	case PeriodTypeDaily, PeriodTypeWeekly, PeriodTypeMonthly, PeriodTypeQuarterly, PeriodTypeAnnual:
		return true
	}
	return false
}

// ReportCategory represents the category of a report type
type ReportCategory string

// ReportCategory values.
const (
	ReportCategoryAcademic       ReportCategory = "academic"
	ReportCategoryAdministrative ReportCategory = "administrative"
	ReportCategoryFinancial      ReportCategory = "financial"
	ReportCategoryMethodical     ReportCategory = "methodical"
)

// IsValid checks if the category is valid
func (c ReportCategory) IsValid() bool {
	switch c {
	case ReportCategoryAcademic, ReportCategoryAdministrative, ReportCategoryFinancial, ReportCategoryMethodical:
		return true
	}
	return false
}

// ParameterType represents the type of a report parameter
type ParameterType string

// ParameterType values.
const (
	ParameterTypeString      ParameterType = "string"
	ParameterTypeNumber      ParameterType = "number"
	ParameterTypeDate        ParameterType = "date"
	ParameterTypeBoolean     ParameterType = "boolean"
	ParameterTypeSelect      ParameterType = "select"
	ParameterTypeMultiSelect ParameterType = "multiselect"
)

// IsValid checks if the parameter type is valid
func (p ParameterType) IsValid() bool {
	switch p {
	case ParameterTypeString, ParameterTypeNumber, ParameterTypeDate,
		ParameterTypeBoolean, ParameterTypeSelect, ParameterTypeMultiSelect:
		return true
	}
	return false
}

// ReportPermission represents permission level for report access
type ReportPermission string

// ReportPermission values.
const (
	ReportPermissionRead    ReportPermission = "read"
	ReportPermissionWrite   ReportPermission = "write"
	ReportPermissionApprove ReportPermission = "approve"
	ReportPermissionPublish ReportPermission = "publish"
)

// IsValid checks if the permission is valid
func (p ReportPermission) IsValid() bool {
	switch p {
	case ReportPermissionRead, ReportPermissionWrite, ReportPermissionApprove, ReportPermissionPublish:
		return true
	}
	return false
}

// AccessRole represents a role that can access reports
type AccessRole string

// AccessRole values.
const (
	AccessRoleAdmin     AccessRole = "admin"
	AccessRoleSecretary AccessRole = "secretary"
	AccessRoleMethodist AccessRole = "methodist"
	AccessRoleTeacher   AccessRole = "teacher"
	AccessRoleStudent   AccessRole = "student"
)

// IsValid checks if the role is valid
func (r AccessRole) IsValid() bool {
	switch r {
	case AccessRoleAdmin, AccessRoleSecretary, AccessRoleMethodist, AccessRoleTeacher, AccessRoleStudent:
		return true
	}
	return false
}

// DeliveryMethod represents how report subscriptions are delivered
type DeliveryMethod string

// DeliveryMethod values.
const (
	DeliveryMethodEmail        DeliveryMethod = "email"
	DeliveryMethodNotification DeliveryMethod = "notification"
	DeliveryMethodBoth         DeliveryMethod = "both"
)

// IsValid checks if the delivery method is valid
func (d DeliveryMethod) IsValid() bool {
	switch d {
	case DeliveryMethodEmail, DeliveryMethodNotification, DeliveryMethodBoth:
		return true
	}
	return false
}

// GenerationStatus represents the status of report generation
type GenerationStatus string

// GenerationStatus values.
const (
	GenerationStatusStarted   GenerationStatus = "started"
	GenerationStatusCompleted GenerationStatus = "completed"
	GenerationStatusFailed    GenerationStatus = "failed"
)

// IsValid checks if the generation status is valid
func (s GenerationStatus) IsValid() bool {
	switch s {
	case GenerationStatusStarted, GenerationStatusCompleted, GenerationStatusFailed:
		return true
	}
	return false
}
