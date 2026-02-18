package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	playScriptRatingsTableName = "play_script_ratings"
	ratingColumn               = "rating"
)

func init() {
	registerTableName(playScriptRatingsTableName)
}

var playScriptRatingsColumns = []string{
	idColumn,
	playScriptIDColumn,
	belongsToUserColumn,
	ratingColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildPlayScriptRatingsQueries(database string) []*Query {
	switch database {
	case postgres:
		insertColumns := filterForInsert(playScriptRatingsColumns)
		fullSelectColumns := applyToEach(playScriptRatingsColumns, func(_ int, s string) string {
			return fullColumnName(playScriptRatingsTableName, s)
		})

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreatePlayScriptRating",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					playScriptRatingsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveRatingForUserAndPlay",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
					playScriptRatingsTableName,
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn, currentTimeExpression,
					playScriptRatingsTableName, archivedAtColumn,
					playScriptRatingsTableName, playScriptIDColumn, playScriptIDColumn,
					playScriptRatingsTableName, belongsToUserColumn, belongsToUserColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetActiveRatingForUserAndPlay",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					playScriptRatingsTableName,
					playScriptRatingsTableName, archivedAtColumn,
					playScriptRatingsTableName, playScriptIDColumn, playScriptIDColumn,
					playScriptRatingsTableName, belongsToUserColumn, belongsToUserColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetRatingStatsForPlayScript",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	COALESCE(AVG(%s.%s)::double precision, 0) AS average_rating,
	COALESCE(COUNT(%s.%s), 0)::bigint AS rating_count
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					playScriptRatingsTableName, ratingColumn,
					playScriptRatingsTableName, idColumn,
					playScriptRatingsTableName,
					playScriptRatingsTableName, playScriptIDColumn, playScriptIDColumn,
					playScriptRatingsTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetRatingStatsForPlayScripts",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s.%s,
	COALESCE(AVG(%s.%s)::double precision, 0) AS average_rating,
	COALESCE(COUNT(%s.%s), 0)::bigint AS rating_count
FROM %s
WHERE %s.%s = ANY(sqlc.arg(%s)::text[])
	AND %s.%s IS NULL
GROUP BY %s.%s;`,
					playScriptRatingsTableName, playScriptIDColumn,
					playScriptRatingsTableName, ratingColumn,
					playScriptRatingsTableName, idColumn,
					playScriptRatingsTableName,
					playScriptRatingsTableName, playScriptIDColumn, "play_script_ids",
					playScriptRatingsTableName, archivedAtColumn,
					playScriptRatingsTableName, playScriptIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetUserRatingsForPlayScripts",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s.%s,
	%s.%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s = ANY(sqlc.arg(%s)::text[])
	AND %s.%s IS NULL;`,
					playScriptRatingsTableName, playScriptIDColumn,
					playScriptRatingsTableName, ratingColumn,
					playScriptRatingsTableName,
					playScriptRatingsTableName, belongsToUserColumn, belongsToUserColumn,
					playScriptRatingsTableName, playScriptIDColumn, "play_script_ids",
					playScriptRatingsTableName, archivedAtColumn,
				)),
			},
		}
	default:
		return nil
	}
}
