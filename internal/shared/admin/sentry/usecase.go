// Package sentry exposes a read-only admin view of the runtime Sentry
// integration. It does not change Sentry behavior — initSentry stays
// in cmd/server/main.go as the single source of truth — it only
// reflects the current configuration so admins can confirm error
// tracking is wired without reading server logs.
package sentry

import (
	"context"
	"os"
)

// Config is the JSON projection returned by GetConfig. DSN is exposed
// as a boolean only — the raw value is a secret and must never leave
// the server even on an admin-gated endpoint.
type Config struct {
	DSNConfigured    bool    `json:"dsn_configured"`
	Environment      string  `json:"environment"`
	Release          string  `json:"release"`
	TracesSampleRate float64 `json:"traces_sample_rate"`
	TracingEnabled   bool    `json:"tracing_enabled"`
}

// DSNProbe answers "is SENTRY_DSN set right now?" without exposing the
// value. The default implementation reads os.Getenv; tests substitute
// fakes to cover both true/false branches deterministically.
type DSNProbe func() bool

// EnvDSNProbe is the production probe: presence of SENTRY_DSN in the
// process environment.
func EnvDSNProbe() bool {
	return os.Getenv("SENTRY_DSN") != ""
}

// AdminSentryUseCase is the read-only admin view. The constants
// (TracesSampleRate, TracingEnabled) mirror cmd/server/main.go
// initSentry — if initSentry ever moves those values into cfg, this
// use case must follow suit to stay a faithful reflection.
type AdminSentryUseCase struct {
	dsnProbe         DSNProbe
	environment      string
	release          string
	tracesSampleRate float64
	tracingEnabled   bool
}

// NewAdminSentryUseCase builds the view. environment and release come
// from cfg.Environment + cfg.Version (the same pair initSentry passes
// to sentry.Init). Panics on nil probe so misconfigured DI fails at
// construction.
func NewAdminSentryUseCase(dsnProbe DSNProbe, environment, release string) *AdminSentryUseCase {
	if dsnProbe == nil {
		panic("sentry: nil DSNProbe")
	}
	return &AdminSentryUseCase{
		dsnProbe:         dsnProbe,
		environment:      environment,
		release:          release,
		tracesSampleRate: 0.1,  // initSentry constant
		tracingEnabled:   true, // initSentry constant
	}
}

// GetConfig returns the current Sentry runtime configuration snapshot.
// RED stub — returns a zero Config so the handler-level integration
// test in handler_test.go fails until the GREEN commit implements the
// real projection.
func (uc *AdminSentryUseCase) GetConfig(_ context.Context) Config {
	return Config{}
}
