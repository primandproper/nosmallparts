// Package migrations is the one sequence every nosmallparts database is brought
// up by.
//
// Part of the schema is not nosmallparts'. Every platform package the service
// composes owns its tables and renders their DDL for a dialect and a table
// prefix, and primitives-go's migrate package splices that text into the
// numbered sequence as if it were a file. So the identity directory and the
// sign-in stores are created by the same run that creates the script library,
// and nothing platform-owned is copied into this repository. The numbers are
// ours to pick because the sequence is ours: the platform block starts at one,
// and our own files start at twenty, leaving room between.
//
// The platform tables carry no foreign key onto identity_users here, where
// fijo adds them. SQLite cannot add a constraint to a table that exists, and
// nothing in this service deletes a user, so the cascade would guard a path
// nobody walks.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"

	"github.com/primandproper/platform-go/v15/authentication/signin/recoverycodes"
	recoverycodesmigrations "github.com/primandproper/platform-go/v15/authentication/signin/recoverycodes/migrations"
	"github.com/primandproper/platform-go/v15/authentication/signin/refreshtokens"
	refreshtokensmigrations "github.com/primandproper/platform-go/v15/authentication/signin/refreshtokens/migrations"
	"github.com/primandproper/platform-go/v15/identity"
	identitymigrations "github.com/primandproper/platform-go/v15/identity/migrations"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	"github.com/primandproper/primitives-go/v2/database/migrate"
	"github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
)

//go:embed migration_files/*.sql
var rawMigrations embed.FS

// The platform block of the sequence. A version, once shipped, is never
// reassigned: a platform package that grows a schema version lands as a new
// number after the last one here, rendered with its SQLSince.
const (
	identityVersion      uint64 = 1
	refreshTokensVersion uint64 = 2
	recoveryCodesVersion uint64 = 3
)

// Migrator brings a nosmallparts database to the current schema. It is what the
// database client runs on boot when RunMigrations is set.
type Migrator struct {
	schema *migrate.Migrator
}

var _ database.Migrator = (*Migrator)(nil)

// NewMigrator renders every platform schema and assembles the sequence. It
// fails here, not at Migrate, for a migration file that does not parse or a
// generated version that collides with one on disk.
func NewMigrator(logger logging.Logger) (*Migrator, error) {
	files, err := fs.Sub(rawMigrations, "migration_files")
	if err != nil {
		return nil, errors.Wrap(err, "opening migration files")
	}

	generated := []struct {
		render  func() (string, error)
		name    string
		version uint64
	}{
		{version: identityVersion, name: "create_identity_tables", render: renderIdentity},
		{version: refreshTokensVersion, name: "create_refresh_tokens_table", render: renderRefreshTokens},
		{version: recoveryCodesVersion, name: "create_recovery_codes_table", render: renderRecoveryCodes},
	}

	opts := []migrate.Option{migrate.WithLogger(logging.EnsureLogger(logger))}

	for idx := range generated {
		g := &generated[idx]

		body, renderErr := g.render()
		if renderErr != nil {
			return nil, errors.Wrapf(renderErr, "rendering %s", g.name)
		}

		opts = append(opts, migrate.WithGeneratedMigration(g.version, g.name, body))
	}

	schema, err := migrate.New(dialect.SQLite, files, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "building migrator")
	}

	return &Migrator{schema: schema}, nil
}

// Migrate applies every pending migration.
func (m *Migrator) Migrate(ctx context.Context, db *sql.DB) error {
	return m.schema.Migrate(ctx, db)
}

func renderIdentity() (string, error) {
	return identitymigrations.SQL(dialect.SQLite, identity.DefaultTablePrefix)
}

func renderRefreshTokens() (string, error) {
	return refreshtokensmigrations.SQL(dialect.SQLite, refreshtokens.DefaultTablePrefix)
}

func renderRecoveryCodes() (string, error) {
	return recoverycodesmigrations.SQL(dialect.SQLite, recoverycodes.DefaultTablePrefix)
}
