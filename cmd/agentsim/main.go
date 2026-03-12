package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/api"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/config"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/content"
	agentlog "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/log"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/orchestrator"

	// Import scenarios to trigger init() registration
	_ "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/scenario"
)

func main() {
	cfg := config.Parse()
	logger := agentlog.New(cfg.Verbose)

	if cfg.Scenario == "" && !cfg.Continuous {
		fmt.Fprintln(os.Stderr, "Usage: agentsim --scenario <name> | --continuous")
		fmt.Fprintln(os.Stderr, "Scenarios: morning, document-flow, semester-start, workday, meeting, reporting-period, all")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("shutting down gracefully...")
		cancel()
	}()

	// Create API client
	client := api.NewClient(cfg.APIURL)

	// Create agent pool
	pool := agent.CreateDefault(15, 5)
	logger.Info("created agent pool", "total", len(pool.All()))

	// Seed users if requested
	if cfg.SeedUsers {
		if err := seedUsers(ctx, client, pool, logger); err != nil {
			logger.Fatal("failed to seed users", "error", err)
		}
	}

	// Create content generator
	var gen content.Generator
	switch {
	case cfg.NoLLM:
		gen = content.NewTemplateGenerator()
	case cfg.LLMAPIKey != "":
		gen = content.NewLLMGenerator(cfg.LLMProvider, cfg.LLMModel, cfg.LLMAPIKey)
	default:
		logger.Info("no LLM API key provided, falling back to template generator")
		gen = content.NewTemplateGenerator()
	}

	// Create and run orchestrator
	orch := orchestrator.New(cfg, client, pool, gen, logger)

	if cfg.Continuous {
		if err := orch.RunContinuous(ctx); err != nil && ctx.Err() == nil {
			logger.Fatal("continuous mode error", "error", err)
		}
	} else {
		if err := orch.RunScenario(ctx, cfg.Scenario); err != nil && ctx.Err() == nil {
			logger.Fatal("scenario error", "error", err)
		}
	}

	logger.Info("agentsim finished")
}

func seedUsers(ctx context.Context, client *api.Client, pool *agent.Pool, logger *agentlog.Logger) error {
	logger.Info("seeding agent users...")
	for _, a := range pool.All() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := client.Register(ctx, a)
		if err != nil {
			// User might already exist, try logging in
			loginErr := client.Login(ctx, a)
			if loginErr != nil {
				logger.Error("failed to seed agent", "email", a.Email, "register_err", err, "login_err", loginErr)
				continue
			}
		}
		logger.AgentAction(a, "registered", fmt.Sprintf("user_id=%d", a.UserID))
	}
	logger.Info("seeding complete")
	return nil
}
