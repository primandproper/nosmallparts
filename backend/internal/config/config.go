// Package config assembles the application's configuration.
//
// The configuration has two halves. The platform half is platform-go's
// service.Config, embedded so its fields promote onto Config: the service name,
// the observability pillars, and one pointer sub-config per platform subsystem
// (database, identity, sign-in, tokens, cookies, the servers, gRPC-web). A
// subsystem is composed into the process by being present in the config and by
// nothing else; see platform-go's service package for the rule. The
// nosmallparts half is what no platform package owns: who the admins are.
//
// A config that names no database, no identity and no servers is still valid,
// which is what `nosmallparts version` boots with. One that serves has to name
// all of them, and ValidateForServing is that check.
//
// The configuration is built in Go with sensible, zero-dependency defaults so
// the binary boots out of the box: structured slog logging, and tracing,
// metrics and profiling left unconfigured, which primitives-go resolves to noop
// providers. Callers may override the service name and log level (see Options),
// which is how the CLI threads its flags and environment into the platform
// configuration.
//
// Two loaders build on those defaults using primitives-go's config package:
//
//   - Load overlays environment variables (prefixed with EnvVarPrefix) on top of
//     the defaults, so any field can be tuned without a config file — the
//     twelve-factor path most deployments start with.
//   - LoadFromFile decodes a complete JSON config file and then overlays the
//     same environment variables, for deployments that graduate to a checked-in
//     or mounted configuration.
//
// Both share envVarOptions, which wires the app's env var prefix and a debug
// hook that logs every value the parser applies.
//
// The set of variables that overlay can read is not written down here. It is
// derived from the `env:` tags reachable from the configurations constraint and
// generated into the envvars subpackage by `make envvars`.
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v15/service"
	primitivesconfig "github.com/primandproper/primitives-go/v2/config"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	loggingcfg "github.com/primandproper/primitives-go/v2/observability/logging/config"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	profilingnoop "github.com/primandproper/primitives-go/v2/observability/profiling/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// DefaultServiceName is the service name reported by the observability suite
// when the caller does not supply one.
const DefaultServiceName = "nosmallparts"

// EnvVarPrefix is prepended to every environment variable this application
// reads, keeping its configuration in a distinct namespace. For example the
// logging level is read from NOSMALLPARTS_OBSERVABILITY_LOGGING_LEVEL: the prefix
// here, then the nested envPrefix tags on Config and the platform sub-configs.
const EnvVarPrefix = "NOSMALLPARTS_"

// Log level names accepted by Options.LogLevel (case-insensitive).
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// configurations is the constraint every loadable configuration struct in this
// package is a member of. Today that is Config alone; a second one — a worker
// or a migration tool with its own shape — joins it here.
//
// It is what makes `make envvars` complete by construction. The generator walks
// this constraint's members rather than a list somebody remembered to update,
// so a configuration struct cannot become loadable without its environment
// variables appearing in the generated envvars package.
type configurations interface {
	Config
}

// Config is the application configuration.
//
// service.Config is embedded rather than named so that its fields promote: the
// JSON a deployment mounts has "name", "observability" and "database" as
// siblings, and the environment reads NOSMALLPARTS_DATABASE_* rather than
// NOSMALLPARTS_PLATFORM_DATABASE_*. The embed carries no env tag on purpose:
// caarlos0/env reads through an anonymous embedded struct and still composes
// the nested envPrefix tags, and encoding/json promotes the fields on its own.
type Config struct {

	// AdminEmails are the addresses whose holders may delete any script and
	// switch one off. They are matched case-insensitively against the signed-in
	// user's address, on every request, so a change takes effect on restart
	// with no data to migrate.
	AdminEmails []string `env:"ADMIN_EMAILS" json:"adminEmails,omitempty"`

	service.Config

	// Stages is how long the timed parts of a stage take.
	Stages StagesConfig `envPrefix:"STAGES_" json:"stages"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// Options tune the values that most deployments care about. Empty fields fall
// back to the built-in defaults.
type Options struct {
	// ServiceName labels telemetry emitted by the service.
	ServiceName string
	// LogLevel is one of "debug", "info", "warn", or "error" (case-insensitive).
	LogLevel string
}

// New builds a Config from the given options, filling in defaults. The result
// validates cleanly and is ready to hand to observability.Config.NewPillars.
func New(opts Options) *Config {
	serviceName := strings.TrimSpace(opts.ServiceName)
	if serviceName == "" {
		serviceName = DefaultServiceName
	}

	cfg := &Config{}
	cfg.Name = serviceName
	cfg.Observability = observability.Config{
		// slog logging to stdout as structured JSON.
		Logging: loggingcfg.Config{
			Provider:    loggingcfg.ProviderSlog,
			ServiceName: serviceName,
			Level:       levelFromString(opts.LogLevel),
		},
		// Tracing, Metrics, and Profiling are left at their zero values.
		// An empty provider is the documented opt-out into a noop, so the
		// binary stays quiet and dependency-free until one is named.
	}

	return cfg
}

// envVarOptions returns the platform config options shared by every loader in
// this package: the application's env var prefix and a debug hook that logs each
// value the parser applies. Because the hook uses the standard-library slog
// default logger, these lines appear only when that logger is at debug level.
func envVarOptions() []primitivesconfig.Option {
	return []primitivesconfig.Option{
		primitivesconfig.WithPrefix(EnvVarPrefix),
		primitivesconfig.WithOnSet(func(tag string, value any, isDefault bool) {
			slog.Debug("config value set from environment",
				slog.String("tag", tag),
				slog.Any("value", value),
				slog.Bool("isDefault", isDefault),
			)
		}),
	}
}

// Load builds a Config from the given options and then overlays environment
// variables on top of it. The options (typically the CLI's flags) seed the
// defaults; any NOSMALLPARTS_-prefixed environment variable that is set wins over
// them. Fields left unset in the environment keep their default value, so the
// binary still boots with structured slog logging and noop telemetry out of the
// box. The result is validated before it is returned.
func Load(ctx context.Context, opts Options) (*Config, error) {
	cfg := New(opts)

	if err := primitivesconfig.ApplyEnvironmentVariables(cfg, envVarOptions()...); err != nil {
		return nil, fmt.Errorf("applying environment variables: %w", err)
	}

	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validating configuration: %w", err)
	}

	return cfg, nil
}

// LoadFromFile decodes a complete JSON configuration file and then overlays
// environment variables (a set NOSMALLPARTS_ variable wins over the file value).
// Unlike Load, it does not start from the built-in defaults: the file is
// expected to fully specify the config, so use it once a deployment has a real
// config file to mount. The result is validated before it is returned. Note
// that validation is permissive about omissions — an empty logging provider is
// the documented opt-out into noop logging, and a service name is required only
// by the providers that export telemetry — so a sparse file loads rather than
// failing, and takes the zero value for whatever it leaves out.
func LoadFromFile(ctx context.Context, path string) (*Config, error) {
	cfg, err := loadFromJSONFile[Config](ctx, path)
	if err != nil {
		return nil, err
	}

	if err = keepFileDatabaseProvider(path, cfg); err != nil {
		return nil, err
	}

	if err = cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validating configuration: %w", err)
	}

	return cfg, nil
}

// databaseProviderEnvVar is the variable that may still choose the database
// provider over the file.
const databaseProviderEnvVar = EnvVarPrefix + "DATABASE_PROVIDER"

// keepFileDatabaseProvider puts the file's database provider back after the
// environment overlay replaced it.
//
// The overlay assigns a field's envDefault whenever its variable is unset,
// over whatever the file said, and primitives-go's database block defaults its
// provider to postgres. So a file naming sqlite loads as postgres unless the
// deployment also sets the variable, which makes the checked-in config a lie
// about the one setting that decides which database this is. The variable still
// wins when it is set; only its absence stops overriding the file.
func keepFileDatabaseProvider(path string, cfg *Config) error {
	if _, set := os.LookupEnv(databaseProviderEnvVar); set || cfg.Database == nil {
		return nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("rereading the configuration file: %w", err)
	}

	var file struct {
		Database *struct {
			Provider string `json:"provider"`
		} `json:"database"`
	}

	if err = json.Unmarshal(raw, &file); err != nil {
		return fmt.Errorf("rereading the configuration file's database provider: %w", err)
	}

	if file.Database != nil && file.Database.Provider != "" {
		cfg.Database.Provider = file.Database.Provider
	}

	return nil
}

// loadFromJSONFile is the decode every loadable configuration struct goes
// through, and the reason the configurations constraint has teeth: a struct
// this package cannot load through here is a struct whose environment variables
// `make envvars` would not have generated.
func loadFromJSONFile[T configurations](ctx context.Context, path string) (*T, error) {
	cfg, err := primitivesconfig.LoadFromJSONFile[T](ctx, path, envVarOptions()...)
	if err != nil {
		return nil, fmt.Errorf("loading configuration file: %w", err)
	}

	return cfg, nil
}

// ValidateWithContext confirms the assembled configuration is internally
// consistent. It shadows the embedded service.Config's method so that a caller
// reaching for the conventional name validates the whole config rather than
// half of it. The platform half validates itself, which is also what releases
// the sub-configs environment parsing allocated and nobody filled in.
func (c *Config) ValidateWithContext(ctx context.Context) error {
	// A file that names no service takes the default name, as New does.
	if strings.TrimSpace(c.Name) == "" {
		c.Name = DefaultServiceName
	}

	if err := c.Config.ValidateWithContext(ctx); err != nil {
		return err
	}

	if err := c.Stages.ValidateWithContext(ctx); err != nil {
		return err
	}

	for _, email := range c.AdminEmails {
		if strings.TrimSpace(email) == "" {
			return platformerrors.New("adminEmails holds an empty address")
		}
	}

	return nil
}

// ValidateForServing is the stricter check `serve` makes: every platform
// subsystem the service is built from has to be present. It is separate from
// ValidateWithContext because `version` is a valid program over a config that
// names none of them.
func (c *Config) ValidateForServing(ctx context.Context) error {
	if err := c.ValidateWithContext(ctx); err != nil {
		return err
	}

	var missing []string

	for name, present := range map[string]bool{
		"database":   c.Database != nil,
		"identity":   c.Identity != nil,
		"signIn":     c.SignIn != nil,
		"tokens":     c.Tokens != nil,
		"cookies":    c.Cookies != nil,
		"routing":    c.Routing != nil,
		"httpServer": c.HTTPServer != nil,
		"grpcServer": c.GRPCServer != nil,
		"grpcWeb":    c.GRPCWeb != nil,
	} {
		if !present {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		slices.Sort(missing)

		return platformerrors.Newf("serving requires these config blocks, which are absent: %s", strings.Join(missing, ", "))
	}

	return nil
}

// NewPillars builds the observability pillars for the application.
//
// Turning on real telemetry is a config change rather than an edit here: name a
// provider in the JSON or in a NOSMALLPARTS_OBSERVABILITY_*_PROVIDER variable and
// the aggregate constructor wires it, exporters and all.
//
// The unconfigured case is built here instead, and the reason is narrower than
// it looks. observability.Config.NewPillars resolves a pillar naming no provider
// to that pillar's noop, which is the same result — but the tracing sub-config
// announces the decision with an unconditional INFO line on its way there, and
// its two siblings do not. Structured logs go to stdout, which is where the
// version subcommand's output goes too, so delegating unconditionally would put
// a log line in front of every `nosmallparts version` at the default level and
// cost the machine-parseable output it promises. There is no
// setting that silences it; building the three noops directly is what does.
func (c *Config) NewPillars(ctx context.Context) (*observability.Pillars, error) {
	if c.telemetryConfigured() {
		return c.Observability.NewPillars(ctx)
	}

	logger, err := loggingcfg.NewLogger(ctx, &c.Observability.Logging)
	if err != nil {
		return nil, fmt.Errorf("setting up logger: %w", err)
	}

	return &observability.Pillars{
		Logger:          logger,
		TracerProvider:  tracingnoop.NewTracerProvider(),
		MetricsProvider: metricsnoop.NewMetricsProvider(),
		Profiler:        profilingnoop.NewProvider(),
	}, nil
}

// telemetryConfigured reports whether any of the three exporting pillars names a
// provider. Naming one — including "noop" — is a deliberate choice, so it is
// answered by the aggregate constructor and whatever that has to say about it.
func (c *Config) telemetryConfigured() bool {
	return c.Observability.Tracing.Provider != "" ||
		c.Observability.Metrics.Provider != "" ||
		c.Observability.Profiling.Provider != ""
}

// levelFromString maps a human-friendly level name onto a platform log level,
// defaulting to info for empty or unrecognized input.
func levelFromString(s string) logging.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case LevelDebug:
		return logging.DebugLevel
	case LevelWarn, "warning":
		return logging.WarnLevel
	case LevelError:
		return logging.ErrorLevel
	default:
		return logging.InfoLevel
	}
}
