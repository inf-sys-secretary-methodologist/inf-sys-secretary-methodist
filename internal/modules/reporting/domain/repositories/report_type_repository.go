package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

// ReportTypeFilter defines filtering options for report type queries
type ReportTypeFilter struct {
	Category   *domain.ReportCategory
	IsPeriodic *bool
}

// ReportTypeRepository defines the interface for report type persistence
type ReportTypeRepository interface {
	// CRUD operations
	Create(ctx context.Context, reportType *entities.ReportType) error
	Save(ctx context.Context, reportType *entities.ReportType) error
	GetByID(ctx context.Context, id int64) (*entities.ReportType, error)
	GetByCode(ctx context.Context, code string) (*entities.ReportType, error)
	Delete(ctx context.Context, id int64) error

	// Query operations
	List(ctx context.Context, filter ReportTypeFilter, limit, offset int) ([]*entities.ReportType, error)
	Count(ctx context.Context, filter ReportTypeFilter) (int64, error)
	GetByCategory(ctx context.Context, category domain.ReportCategory) ([]*entities.ReportType, error)
	GetPeriodic(ctx context.Context) ([]*entities.ReportType, error)

	// Parameters
	AddParameter(ctx context.Context, param *entities.ReportParameter) error
	UpdateParameter(ctx context.Context, param *entities.ReportParameter) error
	DeleteParameter(ctx context.Context, paramID int64) error
	GetParametersByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportParameter, error)

	// Templates
	AddTemplate(ctx context.Context, template *entities.ReportTemplate) error
	UpdateTemplate(ctx context.Context, template *entities.ReportTemplate) error
	DeleteTemplate(ctx context.Context, templateID int64) error
	GetTemplatesByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportTemplate, error)
	GetDefaultTemplate(ctx context.Context, reportTypeID int64) (*entities.ReportTemplate, error)
	SetDefaultTemplate(ctx context.Context, reportTypeID, templateID int64) error

	// Subscriptions
	Subscribe(ctx context.Context, subscription *entities.ReportSubscription) error
	Unsubscribe(ctx context.Context, reportTypeID, userID int64) error
	GetSubscription(ctx context.Context, reportTypeID, userID int64) (*entities.ReportSubscription, error)
	GetSubscribersByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportSubscription, error)
	GetSubscriptionsByUser(ctx context.Context, userID int64) ([]*entities.ReportSubscription, error)
	UpdateSubscription(ctx context.Context, subscription *entities.ReportSubscription) error
}
