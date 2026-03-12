// Package usecases contains business logic for the users module.
package usecases

import (
	"context"
	"fmt"
	"time"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	authRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// UserUseCase handles user management business logic.
type UserUseCase struct {
	userRepo            authRepos.UserRepository
	userProfileRepo     repositories.UserProfileRepository
	departmentRepo      repositories.DepartmentRepository
	positionRepo        repositories.PositionRepository
	auditLogger         *logging.AuditLogger
	notificationUseCase *notifUsecases.NotificationUseCase
}

// NewUserUseCase creates a new user use case.
func NewUserUseCase(
	userRepo authRepos.UserRepository,
	userProfileRepo repositories.UserProfileRepository,
	departmentRepo repositories.DepartmentRepository,
	positionRepo repositories.PositionRepository,
	auditLogger *logging.AuditLogger,
	notificationUseCase *notifUsecases.NotificationUseCase,
) *UserUseCase {
	return &UserUseCase{
		userRepo:            userRepo,
		userProfileRepo:     userProfileRepo,
		departmentRepo:      departmentRepo,
		positionRepo:        positionRepo,
		auditLogger:         auditLogger,
		notificationUseCase: notificationUseCase,
	}
}

// GetUser retrieves a user by ID with organizational info.
func (uc *UserUseCase) GetUser(ctx context.Context, userID int64) (*entities.UserWithOrg, error) {
	return uc.userProfileRepo.GetProfileByID(ctx, userID)
}

// ListUsers retrieves a paginated list of users.
func (uc *UserUseCase) ListUsers(ctx context.Context, filter *dto.UserListFilter) (*dto.UserListResponse, error) {
	// Set defaults
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Convert DTO filter to repository filter
	repoFilter := &repositories.UserFilter{
		DepartmentID: filter.DepartmentID,
		PositionID:   filter.PositionID,
		Role:         filter.Role,
		Status:       filter.Status,
		Search:       filter.Search,
	}

	users, err := uc.userProfileRepo.ListUsersWithOrg(ctx, repoFilter, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.userProfileRepo.CountUsers(ctx, repoFilter)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.UserListResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateUserProfile updates user's organizational profile.
func (uc *UserUseCase) UpdateUserProfile(ctx context.Context, userID int64, input *dto.UpdateUserProfileInput) error {
	// Verify user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify department exists if provided
	if input.DepartmentID != nil {
		_, err := uc.departmentRepo.GetByID(ctx, *input.DepartmentID)
		if err != nil {
			return err
		}
	}

	// Verify position exists if provided
	if input.PositionID != nil {
		_, err := uc.positionRepo.GetByID(ctx, *input.PositionID)
		if err != nil {
			return err
		}
	}

	err = uc.userProfileRepo.UpdateProfile(ctx, userID, input.DepartmentID, input.PositionID, input.Phone, input.Avatar, input.Bio)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "update", "user_profile", map[string]interface{}{
			"user_id":       userID,
			"department_id": input.DepartmentID,
			"position_id":   input.PositionID,
		})
	}

	return nil
}

// UpdateUserRole updates a user's role.
func (uc *UserUseCase) UpdateUserRole(ctx context.Context, userID int64, input *dto.UpdateUserRoleInput) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	oldRole := user.Role
	user.Role = authDomain.RoleType(input.Role)
	user.UpdatedAt = time.Now()

	err = uc.userRepo.Save(ctx, user)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "role_change", "user", map[string]interface{}{
			"user_id":  userID,
			"old_role": oldRole,
			"new_role": user.Role,
		})
	}

	// Notify user about role change
	if uc.notificationUseCase != nil {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			_ = uc.notificationUseCase.SendSystemNotification(
				context.Background(),
				userID,
				"Изменение роли",
				fmt.Sprintf("Ваша роль изменена на «%s»", input.Role),
			)
		}()
	}

	return nil
}

// UpdateUserStatus updates a user's status.
func (uc *UserUseCase) UpdateUserStatus(ctx context.Context, userID int64, input *dto.UpdateUserStatusInput) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	oldStatus := user.Status

	switch input.Status {
	case "active":
		user.Activate()
	case "inactive":
		user.Deactivate()
	case "blocked":
		user.Block()
	}

	err = uc.userRepo.Save(ctx, user)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "status_change", "user", map[string]interface{}{
			"user_id":    userID,
			"old_status": oldStatus,
			"new_status": user.Status,
		})
	}

	// Notify user about status change
	if uc.notificationUseCase != nil {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			statusNames := map[string]string{
				"active":   "активен",
				"inactive": "неактивен",
				"blocked":  "заблокирован",
			}
			statusName := statusNames[input.Status]
			if statusName == "" {
				statusName = input.Status
			}
			_ = uc.notificationUseCase.SendSystemNotification(
				context.Background(),
				userID,
				"Изменение статуса",
				fmt.Sprintf("Ваш статус изменён на «%s»", statusName),
			)
		}()
	}

	return nil
}

// DeleteUser deletes a user.
func (uc *UserUseCase) DeleteUser(ctx context.Context, userID int64) error {
	// Verify user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	err = uc.userRepo.Delete(ctx, userID)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "delete", "user", map[string]interface{}{
			"user_id": userID,
		})
	}

	return nil
}

// BulkUpdateDepartment assigns multiple users to a department.
func (uc *UserUseCase) BulkUpdateDepartment(ctx context.Context, input *dto.BulkUpdateDepartmentInput) error {
	// Verify department exists if provided
	if input.DepartmentID != nil {
		_, err := uc.departmentRepo.GetByID(ctx, *input.DepartmentID)
		if err != nil {
			return err
		}
	}

	err := uc.userProfileRepo.BulkUpdateDepartment(ctx, input.UserIDs, input.DepartmentID)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "bulk_department_update", "user_profile", map[string]interface{}{
			"user_ids":      input.UserIDs,
			"department_id": input.DepartmentID,
		})
	}

	return nil
}

// BulkUpdatePosition assigns multiple users to a position.
func (uc *UserUseCase) BulkUpdatePosition(ctx context.Context, input *dto.BulkUpdatePositionInput) error {
	// Verify position exists if provided
	if input.PositionID != nil {
		_, err := uc.positionRepo.GetByID(ctx, *input.PositionID)
		if err != nil {
			return err
		}
	}

	err := uc.userProfileRepo.BulkUpdatePosition(ctx, input.UserIDs, input.PositionID)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "bulk_position_update", "user_profile", map[string]interface{}{
			"user_ids":    input.UserIDs,
			"position_id": input.PositionID,
		})
	}

	return nil
}

// GetUsersByDepartment retrieves all users in a department.
func (uc *UserUseCase) GetUsersByDepartment(ctx context.Context, departmentID int64) ([]*entities.UserWithOrg, error) {
	return uc.userProfileRepo.GetUsersByDepartment(ctx, departmentID)
}

// GetUsersByPosition retrieves all users with a specific position.
func (uc *UserUseCase) GetUsersByPosition(ctx context.Context, positionID int64) ([]*entities.UserWithOrg, error) {
	return uc.userProfileRepo.GetUsersByPosition(ctx, positionID)
}

// GetBaseUser retrieves base user info from auth module.
func (uc *UserUseCase) GetBaseUser(ctx context.Context, userID int64) (*authEntities.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}
