package usecases

import (
	"fmt"

	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

// Tag-domain sentinels surfaced by TagUseCase. Wrapped через
// shared domainErrors so the existing response.MapDomainError →
// (404/409) routing fires automatically через errors.Is chain.
//
// v0.156.0 ADR-7 (#266): replace fmt.Errorf Russian-string returns
// (CLAUDE.md gate "UI-strings в usecase запрещены"). UI rendering
// stays in handler / messages package; usecase only signals semantics.
var (
	ErrTagNotFound      = fmt.Errorf("documents: tag %w", domainErrors.ErrNotFound)
	ErrTagAlreadyExists = fmt.Errorf("documents: tag %w", domainErrors.ErrAlreadyExists)
)
