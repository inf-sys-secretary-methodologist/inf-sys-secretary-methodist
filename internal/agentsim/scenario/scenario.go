package scenario

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/api"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/content"
)

// Step is a single step in a scenario.
type Step struct {
	Name   string        // Step name for logs
	Agent  string        // Agent name or role (e.g. "Марина Петровна Соколова" or "academic_secretary")
	Delay  time.Duration // Delay before executing this step
	Action func(ctx context.Context, a *agent.Agent, api *api.Client, state *SharedState, gen content.Generator) error
}

// Scenario is a named sequence of steps.
type Scenario struct {
	Name        string
	Description string
	Steps       []Step
}

// Registry holds all registered scenarios.
var Registry = map[string]*Scenario{}

// Register adds a scenario to the registry.
func Register(s *Scenario) {
	Registry[s.Name] = s
}

// AllNames returns all registered scenario names.
func AllNames() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}
