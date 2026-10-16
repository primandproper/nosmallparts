package cli

import (
	"fmt"

	"github.com/primandproper/nosmallparts/backend/internal/build/api"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"

	"github.com/spf13/cobra"
)

// errNoConfig is a subcommand that needs the configuration running before
// bootstrap built one, which cobra's PersistentPreRunE makes impossible
// outside a test that skips it.
var errNoConfig = platformerrors.New("no configuration was loaded")

// newServeCommand returns the `serve` subcommand, which runs the server until
// the process is signaled.
func (a *application) newServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the server: the gRPC services, gRPC-web and the HTTP routes.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.cfg == nil {
				return errNoConfig
			}

			// The service owns its own observability pillars, built from the
			// same config the CLI built its logger from, so the CLI's logger is
			// for the CLI's own lines and the service's for everything else.
			server, err := api.NewServer(cmd.Context(), a.cfg)
			if err != nil {
				return fmt.Errorf("building the server: %w", err)
			}

			a.log().Info("serving")

			return server.Run(cmd.Context())
		},
	}
}

// newSeedCommand returns the `seed` subcommand, which writes the sample plays
// into the configured database and exits.
func (a *application) newSeedCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Write the built-in sample plays into the database. Safe to run more than once.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.cfg == nil {
				return errNoConfig
			}

			if err := api.Seed(cmd.Context(), a.cfg); err != nil {
				return fmt.Errorf("seeding: %w", err)
			}

			a.log().Info("seeded")

			return nil
		},
	}
}
