package repositories

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

// ReportFilter defines filtering options for report queries
type ReportFilter struct {
	ReportTypeID *int64
	AuthorID     *int64
	Status       *domain.ReportStatus
	PeriodStart  *time.Time
	PeriodEnd    *time.Time
	IsPublic     *bool
	Search       *string
}

// ReportRepository defines the interface for report persistence
type ReportRepository interface {
	// CRUD operations
	Create(ctx context.Context, report *entities.Report) error
	Save(ctx context.Context, report *entities.Report) error
	GetByID(ctx context.Context, id int64) (*entities.Report, error)
	Delete(ctx context.Context, id int64) error

	// Query operations
	List(ctx context.Context, filter ReportFilter, limit, offset int) ([]*entities.Report, error)
	Count(ctx context.Context, filter ReportFilter) (int64, error)
	GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Report, error)
	GetByStatus(ctx context.Context, status domain.ReportStatus, limit, offset int) ([]*entities.Report, error)
	GetByReportType(ctx context.Context, reportTypeID int64, limit, offset int) ([]*entities.Report, error)
	GetPublicReports(ctx context.Context, limit, offset int) ([]*entities.Report, error)

	// Access management
	AddAccess(ctx context.Context, access *entities.ReportAccess) error
	RemoveAccess(ctx context.Context, reportID, accessID int64) error
	GetAccessByReport(ctx context.Context, reportID int64) ([]*entities.ReportAccess, error)
	HasAccess(ctx context.Context, reportID, userID int64, permission domain.ReportPermission) (bool, error)

	// Comments
	AddComment(ctx context.Context, comment *entities.ReportComment) error
	UpdateComment(ctx context.Context, comment *entities.ReportComment) error
	DeleteComment(ctx context.Context, commentID int64) error
	GetCommentsByReport(ctx context.Context, reportID int64) ([]*entities.ReportComment, error)

	// History
	AddHistory(ctx context.Context, history *entities.ReportHistory) error
	GetHistoryByReport(ctx context.Context, reportID int64, limit, offset int) ([]*entities.ReportHistory, error)

	// Generation log
	CreateGenerationLog(ctx context.Context, log *entities.ReportGenerationLog) error
	UpdateGenerationLog(ctx context.Context, log *entities.ReportGenerationLog) error
	GetGenerationLogsByReport(ctx context.Context, reportID int64) ([]*entities.ReportGenerationLog, error)
}
