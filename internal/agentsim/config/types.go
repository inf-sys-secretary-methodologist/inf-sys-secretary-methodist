package config

import "time"

// Config holds all agentsim configuration.
type Config struct {
	APIURL      string
	Scenario    string
	Continuous  bool
	Interval    time.Duration
	Speed       float64
	NoLLM       bool
	LLMProvider string
	LLMModel    string
	LLMAPIKey   string
	SeedUsers   bool
	Verbose     bool
}
