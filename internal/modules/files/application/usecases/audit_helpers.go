// Package usecases — module-local audit helpers shared между file и
// version use cases.
package usecases

import (
	"context"
	"fmt"

	filesDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain"
)

// emitAccessDenied emits a standardized `file_<action>_denied` audit event
// against the provided sink. Mirrors v0.160.0 users denial pattern: stable
// action names + `actor_user_id` / `target_file_id` keys так что analytics
// pivots consistently across modules. Sink == nil is tolerated for the
// graceful-degradation contract (audit pipeline optional).
//
// Closes Tier 2 carry-forward — prior to v0.161.1 the helper lived in two
// places (FileUseCase + VersionUseCase) с identical bodies. Shared free
// function eliminates the duplication without coupling either struct to
// the other.
func emitAccessDenied(
	ctx context.Context,
	sink AuditEventLogger,
	actorID, fileID int64,
	action filesDomain.FileAction,
	reasonSuffix string,
) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, fmt.Sprintf("file_%s_denied", action), "file", map[string]interface{}{
		"actor_user_id":  actorID,
		"target_file_id": fileID,
		"reason":         reasonSuffix,
	})
}
