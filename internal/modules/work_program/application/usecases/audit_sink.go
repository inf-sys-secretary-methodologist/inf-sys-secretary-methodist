// Package usecases contains the application-layer use cases for the
// work_program bounded context. Use cases orchestrate the WorkProgram
// aggregate (domain), the WorkProgramRepository port (this package),
// and the AuditSink port (this file).
package usecases

import "context"

// AuditSink is the narrow port the work_program use cases use to emit
// forensic audit events. The platform AuditLogger (*logging.AuditLogger)
// satisfies this interface structurally, keeping use-case tests free of
// the concrete logger and its side effects.
//
// Defined in this package per the Clean Architecture gate:
// audit-style ports live in the consumer (use-case) package, not in
// domain. Mirrors the curriculum/assignments AuditSink shape exactly
// so the adapter wiring in main.go can satisfy multiple module ports
// with a single concrete logger.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}
