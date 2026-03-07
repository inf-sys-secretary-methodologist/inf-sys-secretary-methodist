package orchestrator

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/api"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/config"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/content"
	agentlog "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/log"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/scenario"
)

// Orchestrator coordinates scenario execution.
type Orchestrator struct {
	cfg    *config.Config
	client *api.Client
	pool   *agent.Pool
	gen    content.Generator
	logger *agentlog.Logger
	runner *scenario.Runner
}

// New creates a new Orchestrator.
func New(cfg *config.Config, client *api.Client, pool *agent.Pool, gen content.Generator, logger *agentlog.Logger) *Orchestrator {
	return &Orchestrator{
		cfg:    cfg,
		client: client,
		pool:   pool,
		gen:    gen,
		logger: logger,
		runner: scenario.NewRunner(client, pool, gen, logger, cfg.Speed),
	}
}

// RunScenario runs a named scenario or "all".
func (o *Orchestrator) RunScenario(ctx context.Context, name string) error {
	if name == "all" {
		return o.runAll(ctx)
	}

	s, ok := scenario.Registry[name]
	if !ok {
		return fmt.Errorf("unknown scenario: %s (available: %v)", name, scenario.AllNames())
	}

	return o.runner.Run(ctx, s)
}

func (o *Orchestrator) runAll(ctx context.Context) error {
	order := []string{
		"morning",
		"document-flow",
		"semester-start",
		"workday",
		"meeting",
		"reporting-period",
	}

	for _, name := range order {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		s, ok := scenario.Registry[name]
		if !ok {
			o.logger.Info("skipping unregistered scenario", "name", name)
			continue
		}
		if err := o.runner.Run(ctx, s); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			o.logger.Error("scenario failed, continuing", "scenario", name, "error", err)
		}
	}
	return nil
}

// RunContinuous runs scenarios in a loop with intervals.
func (o *Orchestrator) RunContinuous(ctx context.Context) error {
	return RunContinuous(ctx, o.runner, o.cfg.Interval, o.logger)
}
