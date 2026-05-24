// Package usecases contains business logic for the users module.
package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	usersDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
)

// UserUseCase handles user management business logic.
type UserUseCase struct {
	userRepo            UserAccountRepository
	userProfileRepo     UserProfileRepository
	departmentRepo      DepartmentRepository
	positionRepo        PositionRepository
	auditSink           AuditSink
	notificationUseCase *notifUsecases.NotificationUseCase
}

// NewUserUseCase creates a new user use case.
func NewUserUseCase(
	userRepo UserAccountRepository,
	userProfileRepo UserProfileRepository,
	departmentRepo DepartmentRepository,
	positionRepo PositionRepository,
	auditSink AuditSink,
	notificationUseCase *notifUsecases.NotificationUseCase,
) *UserUseCase {
	return &UserUseCase{
		userRepo:            userRepo,
		userProfileRepo:     userProfileRepo,
		departmentRepo:      departmentRepo,
		positionRepo:        positionRepo,
		auditSink:           auditSink,
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
//
// Authorization: actor must be the target user (self-edit) OR
// system_admin (override). Closes #283 ADR-1 (TIER 0 profile
// takeover) — see [usersDomain.AuthorizeProfileEdit] for the rule.
//
// Audit row records both actor_user_id and target_user_id so attackers
// remain traceable across legitimate admin override flows.
func (uc *UserUseCase) UpdateUserProfile(
	ctx context.Context,
	actorID int64,
	actorRole authDomain.RoleType,
	targetID int64,
	input *dto.UpdateUserProfileInput,
) error {
	if err := usersDomain.AuthorizeProfileEdit(actorID, targetID, actorRole); err != nil {
		// Audit emit denial: failed cross-edit attempts must NOT
		// vanish from the trail (reviewer T1-3 / #283 ADR-1
		// denial-path audit).
		if uc.auditSink != nil {
			uc.auditSink.LogAuditEvent(ctx, "update_denied", "user_profile", map[string]interface{}{
				"actor_user_id":  actorID,
				"target_user_id": targetID,
				"reason":         "profile_edit_forbidden",
			})
		}
		return err
	}

	// Verify avatar key belongs to the target user's prefix (#283 ADR-3).
	// Empty key clears the avatar — always allowed.
	if input.Avatar != "" {
		if err := usersDomain.ValidateAvatarKey(input.Avatar, targetID); err != nil {
			if uc.auditSink != nil {
				uc.auditSink.LogAuditEvent(ctx, "update_denied", "user_profile", map[string]interface{}{
					"actor_user_id":  actorID,
					"target_user_id": targetID,
					"reason":         "invalid_avatar_key",
				})
			}
			return err
		}
	}

	// Verify user exists
	_, err := uc.userRepo.GetByID(ctx, targetID)
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

	err = uc.userProfileRepo.UpdateProfile(ctx, targetID, input.DepartmentID, input.PositionID, input.Phone, input.Avatar, input.Bio)
	if err != nil {
		return err
	}

	if uc.auditSink != nil {
		uc.auditSink.LogAuditEvent(ctx, "update", "user_profile", map[string]interface{}{
			"actor_user_id":  actorID,
			"target_user_id": targetID,
			"department_id":  input.DepartmentID,
			"position_id":    input.PositionID,
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

	if uc.auditSink != nil {
		uc.auditSink.LogAuditEvent(ctx, "role_change", "user", map[string]interface{}{
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
//
// Carries the same #283 ADR-4 Tier 1 guards as DeleteUser when the
// new status is "inactive" or "blocked", because deactivating or
// blocking the last system_admin produces the identical
// administrative-recovery lockout as deleting them. Activating a
// user (the "active" branch) is non-destructive and skips the
// guards.
//
// Self-status-change to "inactive"/"blocked" is always rejected via
// ErrCannotDeleteSelf — the actor would brick their own session
// the moment the next request hits.
func (uc *UserUseCase) UpdateUserStatus(ctx context.Context, actorID, targetID int64, input *dto.UpdateUserStatusInput) error {
	target, err := uc.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	if input.Status == "inactive" || input.Status == "blocked" {
		adminHeadcount := 0
		if target.Role == authDomain.RoleSystemAdmin {
			adminHeadcount, err = uc.userRepo.CountByRole(ctx, authDomain.RoleSystemAdmin)
			if err != nil {
				return err
			}
		}
		if guardErr := usersDomain.AuthorizeUserDelete(actorID, targetID, target.Role, adminHeadcount); guardErr != nil {
			if uc.auditSink != nil {
				reason := "cannot_delete_self"
				if errors.Is(guardErr, usersDomain.ErrLastAdminProtected) {
					reason = "last_admin_protected"
				}
				uc.auditSink.LogAuditEvent(ctx, "status_change_denied", "user", map[string]interface{}{
					"actor_user_id":  actorID,
					"target_user_id": targetID,
					"new_status":     input.Status,
					"reason":         reason,
				})
			}
			return guardErr
		}
	}

	oldStatus := target.Status

	switch input.Status {
	case "active":
		target.Activate()
	case "inactive":
		target.Deactivate()
	case "blocked":
		target.Block()
	}

	err = uc.userRepo.Save(ctx, target)
	if err != nil {
		return err
	}

	if uc.auditSink != nil {
		uc.auditSink.LogAuditEvent(ctx, "status_change", "user", map[string]interface{}{
			"actor_user_id":  actorID,
			"target_user_id": targetID,
			"old_status":     oldStatus,
			"new_status":     target.Status,
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
				targetID,
				"Изменение статуса",
				fmt.Sprintf("Ваш статус изменён на «%s»", statusName),
			)
		}()
	}

	return nil
}

// DeleteUser deletes a user.
//
// Two guards (#283 ADR-4 Tier 1):
//  1. Self-delete (actorID == targetID) is unconditionally forbidden —
//     no role gets to remove its own account, which would brick the
//     actor's session and leave the system in an inconsistent state.
//  2. Removing the last remaining system_admin is forbidden. The
//     headcount query is conditional: only fired when the target is
//     a system_admin (rare path, no perf hit on the common case).
//
// Audit row records both actor_user_id and target_user_id so deletes
// remain traceable.
func (uc *UserUseCase) DeleteUser(ctx context.Context, actorID, targetID int64) error {
	target, err := uc.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	adminHeadcount := 0
	if target.Role == authDomain.RoleSystemAdmin {
		adminHeadcount, err = uc.userRepo.CountByRole(ctx, authDomain.RoleSystemAdmin)
		if err != nil {
			return err
		}
	}

	if err := usersDomain.AuthorizeUserDelete(actorID, targetID, target.Role, adminHeadcount); err != nil {
		// Audit emit denial (reviewer T1-3).
		if uc.auditSink != nil {
			reason := "cannot_delete_self"
			if errors.Is(err, usersDomain.ErrLastAdminProtected) {
				reason = "last_admin_protected"
			}
			uc.auditSink.LogAuditEvent(ctx, "delete_denied", "user", map[string]interface{}{
				"actor_user_id":  actorID,
				"target_user_id": targetID,
				"reason":         reason,
			})
		}
		return err
	}

	if err := uc.userRepo.Delete(ctx, targetID); err != nil {
		return err
	}

	if uc.auditSink != nil {
		uc.auditSink.LogAuditEvent(ctx, "delete", "user", map[string]interface{}{
			"actor_user_id":  actorID,
			"target_user_id": targetID,
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

	if uc.auditSink != nil {
		uc.auditSink.LogAuditEvent(ctx, "bulk_department_update", "user_profile", map[string]interface{}{
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

	if uc.auditSink != nil {
		uc.auditSink.LogAuditEvent(ctx, "bulk_position_update", "user_profile", map[string]interface{}{
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
