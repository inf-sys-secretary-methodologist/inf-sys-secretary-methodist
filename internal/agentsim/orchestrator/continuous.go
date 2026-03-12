package orchestrator

import (
	"context"
	"math/rand"
	"time"

	agentlog "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/log"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/scenario"
)

// RunContinuous runs random scenarios in a loop until the context is cancelled.
func RunContinuous(ctx context.Context, runner *scenario.Runner, interval time.Duration, logger *agentlog.Logger) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404 -- weak RNG is fine for random scenario scheduling

	names := scenario.AllNames()
	if len(names) == 0 {
		logger.Error("no scenarios registered")
		return nil
	}

	logger.Info("starting continuous mode", "scenarios", len(names), "interval", interval.String())

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Pick a random scenario
		name := names[rng.Intn(len(names))]
		s := scenario.Registry[name]
		if s == nil {
			continue
		}

		if err := runner.Run(ctx, s); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			logger.Error("scenario failed in continuous mode", "scenario", name, "error", err)
		}

		// Wait for interval
		logger.Info("waiting before next scenario", "interval", interval.String())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}
