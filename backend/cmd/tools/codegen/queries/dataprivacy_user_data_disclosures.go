package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	userDataDisclosuresTableName = "user_data_disclosures"
	statusColumn                 = "status"
	reportIDColumn               = "report_id"
	expiresAtColumn              = "expires_at"
	completedAtColumn            = "completed_at"
)

func init() {
	registerTableName(userDataDisclosuresTableName)
}

var (
	userDataDisclosuresColumns = []string{
		idColumn,
		belongsToUserColumn,
		statusColumn,
		reportIDColumn,
		expiresAtColumn,
		createdAtColumn,
		lastUpdatedAtColumn,
		completedAtColumn,
		archivedAtColumn,
	}
)

func buildUserDataDisclosuresQueries(database string) []*Query {
	switch database {
	case postgres:
		insertColumns := filterForInsert(userDataDisclosuresColumns, statusColumn, reportIDColumn, completedAtColumn)
		fullSelectColumns := applyToEach(userDataDisclosuresColumns, func(_ int, s string) string {
			return fullColumnName(userDataDisclosuresTableName, s)
		})

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreateUserDataDisclosure",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					userDataDisclosuresTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetUserDataDisclosure",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					userDataDisclosuresTableName,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetUserDataDisclosuresForUser",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(user_id)%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(
						userDataDisclosuresTableName,
						false,
						true,
						nil,
						fmt.Sprintf("%s.%s = sqlc.arg(user_id)", userDataDisclosuresTableName, belongsToUserColumn),
					),
					buildTotalCountSelect(
						userDataDisclosuresTableName,
						true,
						nil,
						fmt.Sprintf("%s.%s = sqlc.arg(user_id)", userDataDisclosuresTableName, belongsToUserColumn),
					),
					userDataDisclosuresTableName,
					userDataDisclosuresTableName, archivedAtColumn,
					userDataDisclosuresTableName, belongsToUserColumn,
					buildFilterConditions(userDataDisclosuresTableName, false, false),
					buildCursorLimitClause(userDataDisclosuresTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MarkUserDataDisclosureCompleted",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = 'completed',
	%s = sqlc.arg(%s),
	%s = %s,
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName,
					statusColumn,
					reportIDColumn, reportIDColumn,
					completedAtColumn, currentTimeExpression,
					lastUpdatedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MarkUserDataDisclosureFailed",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = 'failed',
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName,
					statusColumn,
					lastUpdatedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MarkUserDataDisclosureProcessing",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = 'processing',
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName,
					statusColumn,
					lastUpdatedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveUserDataDisclosure",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName, archivedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
		}
	default:
		return nil
	}
}
