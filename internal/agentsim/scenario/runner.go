package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/api"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/content"
	agentlog "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/log"
)

// Runner executes scenario steps.
type Runner struct {
	client *api.Client
	pool   *agent.Pool
	gen    content.Generator
	logger *agentlog.Logger
	speed  float64
}

// NewRunner creates a new scenario runner.
func NewRunner(client *api.Client, pool *agent.Pool, gen content.Generator, logger *agentlog.Logger, speed float64) *Runner {
	return &Runner{
		client: client,
		pool:   pool,
		gen:    gen,
		logger: logger,
		speed:  speed,
	}
}

// Run executes a scenario.
func (r *Runner) Run(ctx context.Context, s *Scenario) error {
	r.logger.ScenarioStart(s.Name)
	start := time.Now()
	state := NewSharedState()

	for _, step := range s.Steps {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Apply delay with speed multiplier
		if step.Delay > 0 {
			scaled := time.Duration(float64(step.Delay) * r.speed)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(scaled):
			}
		}

		// Resolve agent
		a := r.resolveAgent(step.Agent)
		if a == nil {
			r.logger.StepError(step.Name, fmt.Errorf("agent not found: %s", step.Agent))
			continue
		}

		r.logger.StepStart(step.Name, a.ShortName())

		// Ensure agent is logged in
		if err := r.client.EnsureAuthenticated(ctx, a); err != nil {
			r.logger.StepError(step.Name, fmt.Errorf("auth failed for %s: %w", a.Email, err))
			continue
		}

		// Execute step action
		if err := step.Action(ctx, a, r.client, state, r.gen); err != nil {
			r.logger.StepError(step.Name, err)
			// Continue with next step
		}
	}

	r.logger.ScenarioEnd(s.Name, time.Since(start))
	return nil
}

// resolveAgent finds an agent by name or role.
func (r *Runner) resolveAgent(nameOrRole string) *agent.Agent {
	if a := r.pool.ByName(nameOrRole); a != nil {
		return a
	}
	if a := r.pool.ByEmail(nameOrRole); a != nil {
		return a
	}
	if a := r.pool.FirstByRole(nameOrRole); a != nil {
		return a
	}
	return nil
}
