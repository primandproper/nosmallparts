package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	/* #nosec G101 */
	oauth2ClientTokensTableName = "oauth2_client_tokens"
	codeColumn                  = "code"
	accessColumn                = "access"
	refreshColumn               = "refresh"
	codeExpiresAtColumn         = "code_expires_at"
	accessExpiresAtColumn       = "access_expires_at"
	refreshExpiresAtColumn      = "refresh_expires_at"
)

func init() {
	registerTableName(oauth2ClientTokensTableName)
}

/* #nosec G101 */
var oauth2ClientTokensColumns = []string{
	idColumn,
	"client_id",
	belongsToUserColumn,
	"redirect_uri",
	codeColumn,
	"code_challenge",
	"code_challenge_method",
	"code_created_at",
	codeExpiresAtColumn,
	accessColumn,
	"access_created_at",
	accessExpiresAtColumn,
	refreshColumn,
	"refresh_created_at",
	refreshExpiresAtColumn,
}

func buildOAuth2ClientTokensQueries(database string) []*Query {
	switch database {
	case postgres:

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "DeleteOAuth2ClientTokenByAccess",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s WHERE %s = sqlc.arg(%s);`, oauth2ClientTokensTableName, accessColumn, accessColumn)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "DeleteOAuth2ClientTokenByCode",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s WHERE %s = sqlc.arg(%s);`, oauth2ClientTokensTableName, codeColumn, codeColumn)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "DeleteOAuth2ClientTokenByRefresh",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s WHERE %s = sqlc.arg(%s);`, oauth2ClientTokensTableName, refreshColumn, refreshColumn)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateOAuth2ClientToken",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					oauth2ClientTokensTableName,
					strings.Join(oauth2ClientTokensColumns, ",\n\t"),
					strings.Join(applyToEach(oauth2ClientTokensColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckOAuth2ClientTokenExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					oauth2ClientTokensTableName, idColumn,
					oauth2ClientTokensTableName,
					oauth2ClientTokensTableName, archivedAtColumn,
					oauth2ClientTokensTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetOAuth2ClientTokenByAccess",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s);`,
					strings.Join(applyToEach(oauth2ClientTokensColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", oauth2ClientTokensTableName, s)
					}), ",\n\t"),
					oauth2ClientTokensTableName,
					oauth2ClientTokensTableName, accessColumn, accessColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetOAuth2ClientTokenByCode",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s);`,
					strings.Join(applyToEach(oauth2ClientTokensColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", oauth2ClientTokensTableName, s)
					}), ",\n\t"),
					oauth2ClientTokensTableName,
					oauth2ClientTokensTableName, codeColumn, codeColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetOAuth2ClientTokenByRefresh",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s);`,
					strings.Join(applyToEach(oauth2ClientTokensColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", oauth2ClientTokensTableName, s)
					}), ",\n\t"),
					oauth2ClientTokensTableName,
					oauth2ClientTokensTableName, refreshColumn, refreshColumn,
				)),
			},
		}
	default:
		return nil
	}
}
