package llm

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
)

// Config configures the OpenAI-compatible draft generator. BaseURL /
// APIKey / Model are provider-agnostic (OpenRouter by default, but any
// OpenAI-compatible endpoint works by changing these three).
type Config struct {
	BaseURL     string
	APIKey      string
	Model       string
	Timeout     time.Duration
	Temperature float64
	MaxTokens   int
}

// Generator is an OpenAI-compatible draft generator. STUB — the
// behavior lands in the GREEN commit; this compiles so the RED test
// suite can run and fail.
type Generator struct{}

// compile-time check that the adapter satisfies the application port.
var _ usecases.DraftGenerator = (*Generator)(nil)

// NewGenerator is the STUB constructor (no wiring yet).
func NewGenerator(cfg Config) *Generator { return &Generator{} }

// GenerateDraft is the STUB entrypoint — always returns not-implemented.
func (g *Generator) GenerateDraft(ctx context.Context, req usecases.DraftRequest) (usecases.DraftResult, error) {
	return usecases.DraftResult{}, errors.New("work_program/llm: GenerateDraft not implemented")
}
