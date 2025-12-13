// Package dto contains data transfer objects for the users module.
package dto

// UserCSVRow represents a single user row in CSV format.
type UserCSVRow struct {
	Email          string `csv:"email"`
	Name           string `csv:"name"`
	Role           string `csv:"role"`
	Status         string `csv:"status"`
	Phone          string `csv:"phone"`
	DepartmentCode string `csv:"department_code"`
	PositionCode   string `csv:"position_code"`
}

// DepartmentCSVRow represents a single department row in CSV format.
type DepartmentCSVRow struct {
	Code        string `csv:"code"`
	Name        string `csv:"name"`
	Description string `csv:"description"`
	ParentCode  string `csv:"parent_code"`
	IsActive    bool   `csv:"is_active"`
}

// PositionCSVRow represents a single position row in CSV format.
type PositionCSVRow struct {
	Code        string `csv:"code"`
	Name        string `csv:"name"`
	Description string `csv:"description"`
	Level       int    `csv:"level"`
	IsActive    bool   `csv:"is_active"`
}

// CSVImportResult represents the result of a CSV import operation.
type CSVImportResult struct {
	TotalRows    int      `json:"total_rows"`
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors,omitempty"`
}
