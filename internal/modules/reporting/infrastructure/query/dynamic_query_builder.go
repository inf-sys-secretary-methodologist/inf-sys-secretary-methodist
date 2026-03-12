package query

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

// DataSourceConfig defines the database table and column mappings for a data source
type DataSourceConfig struct {
	TableName      string
	ColumnMappings map[string]string // field name -> actual column expression
	JoinClauses    []string          // Additional JOIN clauses if needed
}

// DynamicQueryBuilder builds and executes dynamic SQL queries for custom reports
type DynamicQueryBuilder struct {
	db            *sql.DB
	sourceConfigs map[entities.DataSourceType]DataSourceConfig
}

// NewDynamicQueryBuilder creates a new DynamicQueryBuilder
func NewDynamicQueryBuilder(db *sql.DB) *DynamicQueryBuilder {
	builder := &DynamicQueryBuilder{
		db:            db,
		sourceConfigs: make(map[entities.DataSourceType]DataSourceConfig),
	}

	// Configure data sources
	builder.configureDataSources()

	return builder
}

// configureDataSources sets up the mappings for all data sources
func (b *DynamicQueryBuilder) configureDataSources() {
	// Documents data source
	b.sourceConfigs[entities.DataSourceDocuments] = DataSourceConfig{
		TableName: "documents d",
		ColumnMappings: map[string]string{
			"id":          "d.id",
			"name":        "d.name",
			"category":    "d.category",
			"status":      "d.status",
			"size":        "d.file_size",
			"created_at":  "d.created_at",
			"updated_at":  "d.updated_at",
			"author_name": "COALESCE(u.first_name || ' ' || u.last_name, 'Unknown')",
			"tags":        "COALESCE((SELECT string_agg(t.name, ', ') FROM document_tags dt JOIN tags t ON dt.tag_id = t.id WHERE dt.document_id = d.id), '')",
		},
		JoinClauses: []string{
			"LEFT JOIN users u ON d.author_id = u.id",
		},
	}

	// Users data source
	b.sourceConfigs[entities.DataSourceUsers] = DataSourceConfig{
		TableName: "users u",
		ColumnMappings: map[string]string{
			"id":         "u.id",
			"name":       "COALESCE(u.first_name || ' ' || u.last_name, u.email)",
			"email":      "u.email",
			"role":       "u.role",
			"department": "COALESCE(u.department, '')",
			"created_at": "u.created_at",
			"is_active":  "u.is_active",
		},
	}

	// Events data source
	b.sourceConfigs[entities.DataSourceEvents] = DataSourceConfig{
		TableName: "events e",
		ColumnMappings: map[string]string{
			"id":         "e.id",
			"title":      "e.title",
			"type":       "e.event_type",
			"start_time": "e.start_time",
			"end_time":   "e.end_time",
			"location":   "COALESCE(e.location, '')",
			"organizer":  "COALESCE(ou.first_name || ' ' || ou.last_name, 'Unknown')",
		},
		JoinClauses: []string{
			"LEFT JOIN users ou ON e.organizer_id = ou.id",
		},
	}

	// Tasks data source
	b.sourceConfigs[entities.DataSourceTasks] = DataSourceConfig{
		TableName: "tasks t",
		ColumnMappings: map[string]string{
			"id":         "t.id",
			"title":      "t.title",
			"status":     "t.status",
			"priority":   "t.priority",
			"due_date":   "t.due_date",
			"assignee":   "COALESCE(au.first_name || ' ' || au.last_name, 'Unassigned')",
			"created_at": "t.created_at",
		},
		JoinClauses: []string{
			"LEFT JOIN users au ON t.assignee_id = au.id",
		},
	}

	// Students data source
	b.sourceConfigs[entities.DataSourceStudents] = DataSourceConfig{
		TableName: "students s",
		ColumnMappings: map[string]string{
			"id":          "s.id",
			"name":        "s.full_name",
			"group":       "COALESCE(g.name, '')",
			"course":      "s.course",
			"faculty":     "COALESCE(s.faculty, '')",
			"status":      "s.status",
			"enrolled_at": "s.enrolled_at",
		},
		JoinClauses: []string{
			"LEFT JOIN student_groups g ON s.group_id = g.id",
		},
	}
}

// Execute executes the report query and returns results
func (b *DynamicQueryBuilder) Execute(ctx context.Context, report *entities.CustomReport, page, pageSize int) (*entities.ReportExecutionResult, error) {
	config, ok := b.sourceConfigs[report.DataSource]
	if !ok {
		return nil, fmt.Errorf("unsupported data source: %s", report.DataSource)
	}

	// Build SELECT clause
	selectCols := make([]string, 0, len(report.Fields))
	columns := make([]entities.ReportColumn, 0, len(report.Fields))

	for _, field := range report.Fields {
		colExpr, ok := config.ColumnMappings[field.Field.Name]
		if !ok {
			continue // Skip unknown fields
		}

		alias := field.Field.Name
		if field.Alias != "" {
			alias = field.Alias
		}

		// Apply aggregation if specified
		if field.Aggregation != "" && field.Aggregation != entities.AggregationNone {
			switch field.Aggregation {
			case entities.AggregationCount:
				colExpr = fmt.Sprintf("COUNT(%s)", colExpr)
			case entities.AggregationSum:
				colExpr = fmt.Sprintf("SUM(%s)", colExpr)
			case entities.AggregationAvg:
				colExpr = fmt.Sprintf("AVG(%s)", colExpr)
			case entities.AggregationMin:
				colExpr = fmt.Sprintf("MIN(%s)", colExpr)
			case entities.AggregationMax:
				colExpr = fmt.Sprintf("MAX(%s)", colExpr)
			}
		}

		selectCols = append(selectCols, fmt.Sprintf("%s AS %s", colExpr, alias))
		columns = append(columns, entities.ReportColumn{
			Key:   alias,
			Label: field.Field.Label,
		})
	}

	if len(selectCols) == 0 {
		return nil, fmt.Errorf("no valid fields selected")
	}

	// Build FROM clause with JOINs
	fromClause := config.TableName
	if len(config.JoinClauses) > 0 {
		fromClause += " " + strings.Join(config.JoinClauses, " ")
	}

	// Build WHERE clause
	whereClauses, whereArgs := b.buildWhereClause(report.Filters, config)

	// Build GROUP BY clause
	groupByCols := make([]string, 0)
	for _, g := range report.Groupings {
		if colExpr, ok := config.ColumnMappings[g.Field.Name]; ok {
			groupByCols = append(groupByCols, colExpr)
		}
	}

	// Build ORDER BY clause
	orderByCols := make([]string, 0)
	for _, s := range report.Sortings {
		if colExpr, ok := config.ColumnMappings[s.Field.Name]; ok {
			order := "ASC"
			if s.Order == entities.SortOrderDesc {
				order = "DESC"
			}
			orderByCols = append(orderByCols, fmt.Sprintf("%s %s", colExpr, order))
		}
	}

	// Default ordering
	if len(orderByCols) == 0 {
		orderByCols = append(orderByCols, "1 ASC")
	}

	// Build count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", fromClause) // #nosec G201 -- dynamic column/table names from code, not user input
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var totalCount int64
	err := b.db.QueryRowContext(ctx, countQuery, whereArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Build main query
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectCols, ", "), fromClause) // #nosec G201 -- dynamic column/table names from code, not user input
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	if len(groupByCols) > 0 {
		query += " GROUP BY " + strings.Join(groupByCols, ", ")
	}
	query += " ORDER BY " + strings.Join(orderByCols, ", ")

	// Add pagination
	offset := (page - 1) * pageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	// Execute query
	rows, err := b.db.QueryContext(ctx, query, whereArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// Get column names
	colNames, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Read results
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(colNames))
		valuePtrs := make([]interface{}, len(colNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create map for this row
		row := make(map[string]interface{})
		for i, colName := range colNames {
			val := values[i]
			// Convert byte arrays to strings
			if b, ok := val.([]byte); ok {
				row[colName] = string(b)
			} else {
				row[colName] = val
			}
		}

		results = append(results, row)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &entities.ReportExecutionResult{
		Columns:    columns,
		Rows:       results,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// buildWhereClause builds WHERE clause from filters
func (b *DynamicQueryBuilder) buildWhereClause(filters []entities.ReportFilterConfig, config DataSourceConfig) ([]string, []interface{}) {
	clauses := make([]string, 0)
	args := make([]interface{}, 0)
	argIndex := 1

	for _, filter := range filters {
		colExpr, ok := config.ColumnMappings[filter.Field.Name]
		if !ok {
			continue
		}

		var clause string
		switch filter.Operator {
		case entities.FilterEquals:
			clause = fmt.Sprintf("%s = $%d", colExpr, argIndex)
			args = append(args, filter.Value)
			argIndex++
		case entities.FilterNotEquals:
			clause = fmt.Sprintf("%s != $%d", colExpr, argIndex)
			args = append(args, filter.Value)
			argIndex++
		case entities.FilterContains:
			clause = fmt.Sprintf("%s ILIKE $%d", colExpr, argIndex)
			args = append(args, "%"+fmt.Sprint(filter.Value)+"%")
			argIndex++
		case entities.FilterNotContains:
			clause = fmt.Sprintf("%s NOT ILIKE $%d", colExpr, argIndex)
			args = append(args, "%"+fmt.Sprint(filter.Value)+"%")
			argIndex++
		case entities.FilterStartsWith:
			clause = fmt.Sprintf("%s ILIKE $%d", colExpr, argIndex)
			args = append(args, fmt.Sprint(filter.Value)+"%")
			argIndex++
		case entities.FilterEndsWith:
			clause = fmt.Sprintf("%s ILIKE $%d", colExpr, argIndex)
			args = append(args, "%"+fmt.Sprint(filter.Value))
			argIndex++
		case entities.FilterGreaterThan:
			clause = fmt.Sprintf("%s > $%d", colExpr, argIndex)
			args = append(args, filter.Value)
			argIndex++
		case entities.FilterLessThan:
			clause = fmt.Sprintf("%s < $%d", colExpr, argIndex)
			args = append(args, filter.Value)
			argIndex++
		case entities.FilterGreaterOrEqual:
			clause = fmt.Sprintf("%s >= $%d", colExpr, argIndex)
			args = append(args, filter.Value)
			argIndex++
		case entities.FilterLessOrEqual:
			clause = fmt.Sprintf("%s <= $%d", colExpr, argIndex)
			args = append(args, filter.Value)
			argIndex++
		case entities.FilterBetween:
			clause = fmt.Sprintf("%s BETWEEN $%d AND $%d", colExpr, argIndex, argIndex+1)
			args = append(args, filter.Value, filter.Value2)
			argIndex += 2
		case entities.FilterIn:
			// Handle array value
			if arr, ok := filter.Value.([]interface{}); ok {
				placeholders := make([]string, len(arr))
				for i, v := range arr {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				clause = fmt.Sprintf("%s IN (%s)", colExpr, strings.Join(placeholders, ", "))
			}
		case entities.FilterNotIn:
			if arr, ok := filter.Value.([]interface{}); ok {
				placeholders := make([]string, len(arr))
				for i, v := range arr {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				clause = fmt.Sprintf("%s NOT IN (%s)", colExpr, strings.Join(placeholders, ", "))
			}
		case entities.FilterIsNull:
			clause = fmt.Sprintf("%s IS NULL", colExpr)
		case entities.FilterIsNotNull:
			clause = fmt.Sprintf("%s IS NOT NULL", colExpr)
		}

		if clause != "" {
			clauses = append(clauses, clause)
		}
	}

	return clauses, args
}

// Export exports report results to the specified format
func (b *DynamicQueryBuilder) Export(result *entities.ReportExecutionResult, options entities.ExportOptions, reportName string) ([]byte, string, error) {
	switch options.Format {
	case entities.ExportFormatCSV:
		return b.exportCSV(result, options, reportName)
	case entities.ExportFormatXLSX:
		return b.exportXLSX(result, options, reportName)
	case entities.ExportFormatPDF:
		return b.exportPDF(result, options, reportName)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", options.Format)
	}
}

// exportCSV exports to CSV format
func (b *DynamicQueryBuilder) exportCSV(result *entities.ReportExecutionResult, options entities.ExportOptions, reportName string) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if options.IncludeHeaders {
		headers := make([]string, len(result.Columns))
		for i, col := range result.Columns {
			headers[i] = col.Label
		}
		if err := writer.Write(headers); err != nil {
			return nil, "", err
		}
	}

	// Write data rows
	for _, row := range result.Rows {
		record := make([]string, len(result.Columns))
		for i, col := range result.Columns {
			record[i] = formatValue(row[col.Key])
		}
		if err := writer.Write(record); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("%s_%s.csv", sanitizeFilename(reportName), time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// exportXLSX exports to Excel format
func (b *DynamicQueryBuilder) exportXLSX(result *entities.ReportExecutionResult, options entities.ExportOptions, reportName string) ([]byte, string, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheetName := "Report"
	_ = f.SetSheetName("Sheet1", sheetName)

	rowIndex := 1

	// Write header
	if options.IncludeHeaders {
		for i, col := range result.Columns {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIndex)
			_ = f.SetCellValue(sheetName, cell, col.Label)
		}
		rowIndex++
	}

	// Write data rows
	for _, row := range result.Rows {
		for i, col := range result.Columns {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIndex)
			_ = f.SetCellValue(sheetName, cell, row[col.Key])
		}
		rowIndex++
	}

	// Auto-fit columns
	for i := range result.Columns {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheetName, colName, colName, 15)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("%s_%s.xlsx", sanitizeFilename(reportName), time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// exportPDF exports to PDF format
func (b *DynamicQueryBuilder) exportPDF(result *entities.ReportExecutionResult, options entities.ExportOptions, reportName string) ([]byte, string, error) {
	// Determine page size and orientation
	pageSize := "A4"
	if options.PageSize != "" {
		pageSize = options.PageSize
	}

	orientation := "P" // Portrait
	if options.Orientation == "landscape" {
		orientation = "L"
	}

	pdf := gofpdf.New(orientation, "mm", pageSize, "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, reportName)
	pdf.Ln(15)

	// Calculate column widths
	numCols := len(result.Columns)
	if numCols == 0 {
		return nil, "", fmt.Errorf("no columns in report")
	}

	pageWidth, _ := pdf.GetPageSize()
	marginLeft, _, marginRight, _ := pdf.GetMargins()
	usableWidth := pageWidth - marginLeft - marginRight
	colWidth := usableWidth / float64(numCols)

	// Header
	if options.IncludeHeaders {
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(240, 240, 240)
		for _, col := range result.Columns {
			pdf.CellFormat(colWidth, 8, truncateString(col.Label, int(colWidth/2)), "1", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}

	// Data rows
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)
	for _, row := range result.Rows {
		for _, col := range result.Columns {
			value := formatValue(row[col.Key])
			pdf.CellFormat(colWidth, 6, truncateString(value, int(colWidth/2)), "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)

		// Check if we need a new page
		if pdf.GetY() > 270 {
			pdf.AddPage()
			// Reprint header on new page
			if options.IncludeHeaders {
				pdf.SetFont("Arial", "B", 10)
				pdf.SetFillColor(240, 240, 240)
				for _, col := range result.Columns {
					pdf.CellFormat(colWidth, 8, truncateString(col.Label, int(colWidth/2)), "1", 0, "L", true, 0, "")
				}
				pdf.Ln(-1)
				pdf.SetFont("Arial", "", 9)
			}
		}
	}

	// Footer with generation timestamp
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("%s_%s.pdf", sanitizeFilename(reportName), time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// formatValue formats a value for export
func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	case bool:
		if val {
			return "Yes"
		}
		return "No"
	default:
		return fmt.Sprint(val)
	}
}

// sanitizeFilename removes or replaces invalid characters from filename
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GetAvailableFields returns the available fields for a data source
func (b *DynamicQueryBuilder) GetAvailableFields(dataSource entities.DataSourceType) []entities.ReportField {
	fields := make([]entities.ReportField, 0)

	switch dataSource {
	case entities.DataSourceDocuments:
		fields = []entities.ReportField{
			{ID: "id", Name: "id", Label: "ID", Type: entities.FieldTypeNumber, Source: dataSource},
			{ID: "name", Name: "name", Label: "Название", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "category", Name: "category", Label: "Категория", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "status", Name: "status", Label: "Статус", Type: entities.FieldTypeEnum, Source: dataSource, EnumValues: []string{"draft", "pending", "approved", "rejected"}},
			{ID: "size", Name: "size", Label: "Размер (байт)", Type: entities.FieldTypeNumber, Source: dataSource},
			{ID: "created_at", Name: "created_at", Label: "Дата создания", Type: entities.FieldTypeDate, Source: dataSource},
			{ID: "updated_at", Name: "updated_at", Label: "Дата обновления", Type: entities.FieldTypeDate, Source: dataSource},
			{ID: "author_name", Name: "author_name", Label: "Автор", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "tags", Name: "tags", Label: "Теги", Type: entities.FieldTypeString, Source: dataSource},
		}
	case entities.DataSourceUsers:
		fields = []entities.ReportField{
			{ID: "id", Name: "id", Label: "ID", Type: entities.FieldTypeNumber, Source: dataSource},
			{ID: "name", Name: "name", Label: "Имя", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "email", Name: "email", Label: "Email", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "role", Name: "role", Label: "Роль", Type: entities.FieldTypeEnum, Source: dataSource, EnumValues: []string{"admin", "user", "moderator"}},
			{ID: "department", Name: "department", Label: "Подразделение", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "created_at", Name: "created_at", Label: "Дата регистрации", Type: entities.FieldTypeDate, Source: dataSource},
			{ID: "is_active", Name: "is_active", Label: "Активен", Type: entities.FieldTypeBoolean, Source: dataSource},
		}
	case entities.DataSourceEvents:
		fields = []entities.ReportField{
			{ID: "id", Name: "id", Label: "ID", Type: entities.FieldTypeNumber, Source: dataSource},
			{ID: "title", Name: "title", Label: "Название", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "type", Name: "type", Label: "Тип", Type: entities.FieldTypeEnum, Source: dataSource, EnumValues: []string{"meeting", "lecture", "seminar", "exam"}},
			{ID: "start_time", Name: "start_time", Label: "Начало", Type: entities.FieldTypeDate, Source: dataSource},
			{ID: "end_time", Name: "end_time", Label: "Окончание", Type: entities.FieldTypeDate, Source: dataSource},
			{ID: "location", Name: "location", Label: "Место", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "organizer", Name: "organizer", Label: "Организатор", Type: entities.FieldTypeString, Source: dataSource},
		}
	case entities.DataSourceTasks:
		fields = []entities.ReportField{
			{ID: "id", Name: "id", Label: "ID", Type: entities.FieldTypeNumber, Source: dataSource},
			{ID: "title", Name: "title", Label: "Название", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "status", Name: "status", Label: "Статус", Type: entities.FieldTypeEnum, Source: dataSource, EnumValues: []string{"pending", "in_progress", "completed", "cancelled"}},
			{ID: "priority", Name: "priority", Label: "Приоритет", Type: entities.FieldTypeEnum, Source: dataSource, EnumValues: []string{"low", "medium", "high", "critical"}},
			{ID: "due_date", Name: "due_date", Label: "Срок выполнения", Type: entities.FieldTypeDate, Source: dataSource},
			{ID: "assignee", Name: "assignee", Label: "Исполнитель", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "created_at", Name: "created_at", Label: "Дата создания", Type: entities.FieldTypeDate, Source: dataSource},
		}
	case entities.DataSourceStudents:
		fields = []entities.ReportField{
			{ID: "id", Name: "id", Label: "ID", Type: entities.FieldTypeNumber, Source: dataSource},
			{ID: "name", Name: "name", Label: "ФИО", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "group", Name: "group", Label: "Группа", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "course", Name: "course", Label: "Курс", Type: entities.FieldTypeNumber, Source: dataSource},
			{ID: "faculty", Name: "faculty", Label: "Факультет", Type: entities.FieldTypeString, Source: dataSource},
			{ID: "status", Name: "status", Label: "Статус", Type: entities.FieldTypeEnum, Source: dataSource, EnumValues: []string{"active", "graduated", "expelled", "academic_leave"}},
			{ID: "enrolled_at", Name: "enrolled_at", Label: "Дата зачисления", Type: entities.FieldTypeDate, Source: dataSource},
		}
	}

	return fields
}
