package entities

import "time"

// PermissionLevel represents the level of access to a document
type PermissionLevel string

const (
	PermissionRead   PermissionLevel = "read"
	PermissionWrite  PermissionLevel = "write"
	PermissionDelete PermissionLevel = "delete"
	PermissionAdmin  PermissionLevel = "admin"
)

// UserRole represents the role that can be granted permissions
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleSecretary UserRole = "secretary"
	RoleMethodist UserRole = "methodist"
	RoleTeacher   UserRole = "teacher"
	RoleStudent   UserRole = "student"
)

// DocumentPermission represents a permission granted to a user or role for a document
type DocumentPermission struct {
	ID         int64           `db:"id" json:"id"`
	DocumentID int64           `db:"document_id" json:"document_id"`
	UserID     *int64          `db:"user_id" json:"user_id,omitempty"`
	Role       *UserRole       `db:"role" json:"role,omitempty"`
	Permission PermissionLevel `db:"permission" json:"permission"`
	GrantedBy  *int64          `db:"granted_by" json:"granted_by,omitempty"`
	ExpiresAt  *time.Time      `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`

	// Populated via JOIN
	UserName      *string `db:"-" json:"user_name,omitempty"`
	UserEmail     *string `db:"-" json:"user_email,omitempty"`
	GrantedByName *string `db:"-" json:"granted_by_name,omitempty"`
}

// IsExpired checks if the permission has expired
func (p *DocumentPermission) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}

// IsValid checks if the permission is valid (not expired and has user or role)
func (p *DocumentPermission) IsValid() bool {
	if p.IsExpired() {
		return false
	}
	return p.UserID != nil || p.Role != nil
}

// CanRead checks if permission allows reading
func (p *DocumentPermission) CanRead() bool {
	return p.Permission == PermissionRead ||
		p.Permission == PermissionWrite ||
		p.Permission == PermissionDelete ||
		p.Permission == PermissionAdmin
}

// CanWrite checks if permission allows writing
func (p *DocumentPermission) CanWrite() bool {
	return p.Permission == PermissionWrite ||
		p.Permission == PermissionDelete ||
		p.Permission == PermissionAdmin
}

// CanDelete checks if permission allows deleting
func (p *DocumentPermission) CanDelete() bool {
	return p.Permission == PermissionDelete ||
		p.Permission == PermissionAdmin
}

// IsAdmin checks if permission is admin level
func (p *DocumentPermission) IsAdmin() bool {
	return p.Permission == PermissionAdmin
}
