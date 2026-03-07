package config

import (
	"flag"
	"os"
	"time"
)

// Parse reads configuration from command-line flags and environment variables.
func Parse() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.APIURL, "api-url", "http://localhost:8080", "Server API URL")
	flag.StringVar(&cfg.Scenario, "scenario", "", "Scenario name to run (morning, document-flow, semester-start, workday, meeting, reporting-period, all)")
	flag.BoolVar(&cfg.Continuous, "continuous", false, "Run in continuous mode")
	flag.DurationVar(&cfg.Interval, "interval", 5*time.Minute, "Pause between scenarios in continuous mode")
	flag.Float64Var(&cfg.Speed, "speed", 0.5, "Speed multiplier: 1.0=realtime, 0.1=fast")
	flag.BoolVar(&cfg.NoLLM, "no-llm", false, "Disable LLM, use templates only")
	flag.StringVar(&cfg.LLMProvider, "llm-provider", "anthropic", "LLM provider: anthropic|openai")
	flag.StringVar(&cfg.LLMModel, "llm-model", "claude-haiku-4-5-20251001", "LLM model name")
	flag.StringVar(&cfg.LLMAPIKey, "llm-api-key", "", "LLM API key (or env AGENTSIM_LLM_API_KEY)")
	flag.BoolVar(&cfg.SeedUsers, "seed-users", false, "Register agent users on startup")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Verbose logging")

	flag.Parse()

	// Env variable fallbacks
	if cfg.APIURL == "http://localhost:8080" {
		if envURL := os.Getenv("AGENTSIM_API_URL"); envURL != "" {
			cfg.APIURL = envURL
		}
	}

	if cfg.LLMAPIKey == "" {
		cfg.LLMAPIKey = os.Getenv("AGENTSIM_LLM_API_KEY")
	}

	if cfg.Speed <= 0 {
		cfg.Speed = 0.5
	}

	return cfg
}

// ScaledDuration applies the speed multiplier to a duration.
func (c *Config) ScaledDuration(d time.Duration) time.Duration {
	return time.Duration(float64(d) * c.Speed)
}
