package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
)

// ReportAccess represents access permissions for a report
type ReportAccess struct {
	ID         int64                   `json:"id"`
	ReportID   int64                   `json:"report_id"`
	UserID     *int64                  `json:"user_id,omitempty"`
	Role       *domain.AccessRole      `json:"role,omitempty"`
	Permission domain.ReportPermission `json:"permission"`
	GrantedBy  *int64                  `json:"granted_by,omitempty"`
	CreatedAt  time.Time               `json:"created_at"`
}

// NewReportAccessForUser creates access for a specific user
func NewReportAccessForUser(reportID, userID int64, permission domain.ReportPermission, grantedBy *int64) *ReportAccess {
	return &ReportAccess{
		ReportID:   reportID,
		UserID:     &userID,
		Permission: permission,
		GrantedBy:  grantedBy,
		CreatedAt:  time.Now(),
	}
}

// NewReportAccessForRole creates access for a role
func NewReportAccessForRole(reportID int64, role domain.AccessRole, permission domain.ReportPermission, grantedBy *int64) *ReportAccess {
	return &ReportAccess{
		ReportID:   reportID,
		Role:       &role,
		Permission: permission,
		GrantedBy:  grantedBy,
		CreatedAt:  time.Now(),
	}
}

// IsForUser checks if access is granted to a specific user
func (ra *ReportAccess) IsForUser() bool {
	return ra.UserID != nil
}

// IsForRole checks if access is granted to a role
func (ra *ReportAccess) IsForRole() bool {
	return ra.Role != nil
}

// ReportComment represents a comment on a report
type ReportComment struct {
	ID        int64     `json:"id"`
	ReportID  int64     `json:"report_id"`
	AuthorID  int64     `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewReportComment creates a new comment
func NewReportComment(reportID, authorID int64, content string) *ReportComment {
	now := time.Now()
	return &ReportComment{
		ReportID:  reportID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the comment content
func (rc *ReportComment) Update(content string) {
	rc.Content = content
	rc.UpdatedAt = time.Now()
}

// ReportSubscription represents a user subscription to a report type
type ReportSubscription struct {
	ID             int64                 `json:"id"`
	ReportTypeID   int64                 `json:"report_type_id"`
	UserID         int64                 `json:"user_id"`
	DeliveryMethod domain.DeliveryMethod `json:"delivery_method"`
	IsActive       bool                  `json:"is_active"`
	CreatedAt      time.Time             `json:"created_at"`
}

// NewReportSubscription creates a new subscription
func NewReportSubscription(reportTypeID, userID int64, deliveryMethod domain.DeliveryMethod) *ReportSubscription {
	return &ReportSubscription{
		ReportTypeID:   reportTypeID,
		UserID:         userID,
		DeliveryMethod: deliveryMethod,
		IsActive:       true,
		CreatedAt:      time.Now(),
	}
}

// Activate activates the subscription
func (rs *ReportSubscription) Activate() {
	rs.IsActive = true
}

// Deactivate deactivates the subscription
func (rs *ReportSubscription) Deactivate() {
	rs.IsActive = false
}

// SetDeliveryMethod changes the delivery method
func (rs *ReportSubscription) SetDeliveryMethod(method domain.DeliveryMethod) {
	rs.DeliveryMethod = method
}
