//go:build integration || pgtap

// Package testutil provides integration test orchestration.
package testutil

// Option configures integration test behavior.
type Option func(*IntegrationTestOptions)

// IntegrationTestOptions controls the test container lifecycle.
type IntegrationTestOptions struct {
	RunMigrations  bool
	CreateSnapshot bool
	SkipIfNoDocker bool
}

// WithMigrations runs database migrations after container starts.
func WithMigrations() Option {
	return func(o *IntegrationTestOptions) {
		o.RunMigrations = true
	}
}

// WithSnapshot creates a database snapshot after migrations for fast restore.
func WithSnapshot() Option {
	return func(o *IntegrationTestOptions) {
		o.CreateSnapshot = true
	}
}

// SkipIfNoDocker skips tests if Docker/Podman is not available.
func SkipIfNoDocker() Option {
	return func(o *IntegrationTestOptions) {
		o.SkipIfNoDocker = true
	}
}

// applyOptions applies all options to the config.
func applyOptions(opts []Option) *IntegrationTestOptions {
	cfg := &IntegrationTestOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
