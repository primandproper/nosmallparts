package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	playScriptsTableName = "play_scripts"
	titleColumn          = "title"
	authorColumn         = "author"
	isPublicColumn       = "is_public"
	precededByIDColumn   = "preceded_by_id"
	precedesIDColumn     = "precedes_id"
)

func init() {
	registerTableName(playScriptsTableName)
}

var playScriptsColumns = []string{
	idColumn,
	titleColumn,
	descriptionColumn,
	authorColumn,
	isPublicColumn,
	precededByIDColumn,
	precedesIDColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	createdByUserColumn,
	belongsToAccountColumn,
}

func buildPlayScriptsQueries(database string) []*Query {
	switch database {
	case postgres:
		insertColumns := filterForInsert(playScriptsColumns)
		fullSelectColumns := applyToEach(playScriptsColumns, func(_ int, s string) string {
			return fullColumnName(playScriptsTableName, s)
		})

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreatePlayScript",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					playScriptsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdatePlayScript",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					playScriptsTableName,
					strings.Join(applyToEach(filterForUpdate(playScriptsColumns, createdByUserColumn, belongsToAccountColumn), func(_ int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchivePlayScript",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					playScriptsTableName,
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetPlayScript",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					playScriptsTableName,
					playScriptsTableName, archivedAtColumn,
					playScriptsTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckPlayScriptExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					playScriptsTableName, idColumn,
					playScriptsTableName,
					playScriptsTableName, archivedAtColumn,
					playScriptsTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetPlayScripts",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(playScriptsTableName, true, true, nil),
					buildTotalCountSelect(playScriptsTableName, true, nil),
					playScriptsTableName,
					playScriptsTableName, archivedAtColumn,
					buildFilterConditions(playScriptsTableName, true, false),
					buildCursorLimitClause(playScriptsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetPlayScriptsForAccount",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(playScriptsTableName, true, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", playScriptsTableName, belongsToAccountColumn, belongsToAccountColumn)),
					buildTotalCountSelect(playScriptsTableName, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", playScriptsTableName, belongsToAccountColumn, belongsToAccountColumn)),
					playScriptsTableName,
					playScriptsTableName, archivedAtColumn,
					playScriptsTableName, belongsToAccountColumn, belongsToAccountColumn,
					buildFilterConditions(playScriptsTableName, true, false, fmt.Sprintf("%s.%s = sqlc.arg(%s)", playScriptsTableName, belongsToAccountColumn, belongsToAccountColumn)),
					buildCursorLimitClause(playScriptsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetPublicPlayScripts",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = TRUE
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(playScriptsTableName, true, true, nil, fmt.Sprintf("%s.%s = TRUE", playScriptsTableName, isPublicColumn)),
					buildTotalCountSelect(playScriptsTableName, true, nil, fmt.Sprintf("%s.%s = TRUE", playScriptsTableName, isPublicColumn)),
					playScriptsTableName,
					playScriptsTableName, archivedAtColumn,
					playScriptsTableName, isPublicColumn,
					buildFilterConditions(playScriptsTableName, true, false, fmt.Sprintf("%s.%s = TRUE", playScriptsTableName, isPublicColumn)),
					buildCursorLimitClause(playScriptsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "SearchPlayScriptsForXParticipants",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND (SELECT COUNT(*) FROM play_script_cast_members WHERE play_script_cast_members.play_script_id = %s.id) = sqlc.arg(participant_count)
	AND (%s.%s = TRUE OR %s.%s = sqlc.arg(belongs_to_account))
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(playScriptsTableName, true, true, nil,
						fmt.Sprintf("(%s.%s = TRUE OR %s.%s = sqlc.arg(belongs_to_account))", playScriptsTableName, isPublicColumn, playScriptsTableName, belongsToAccountColumn),
						"(SELECT COUNT(*) FROM play_script_cast_members WHERE play_script_cast_members.play_script_id = play_scripts.id) = sqlc.arg(participant_count)",
					),
					buildTotalCountSelect(playScriptsTableName, true, nil,
						fmt.Sprintf("(%s.%s = TRUE OR %s.%s = sqlc.arg(belongs_to_account))", playScriptsTableName, isPublicColumn, playScriptsTableName, belongsToAccountColumn),
						"(SELECT COUNT(*) FROM play_script_cast_members WHERE play_script_cast_members.play_script_id = play_scripts.id) = sqlc.arg(participant_count)",
					),
					playScriptsTableName,
					playScriptsTableName, archivedAtColumn,
					playScriptsTableName,
					playScriptsTableName, isPublicColumn,
					playScriptsTableName, belongsToAccountColumn,
					buildFilterConditions(playScriptsTableName, true, false,
						fmt.Sprintf("(%s.%s = TRUE OR %s.%s = sqlc.arg(belongs_to_account))", playScriptsTableName, isPublicColumn, playScriptsTableName, belongsToAccountColumn),
						"(SELECT COUNT(*) FROM play_script_cast_members WHERE play_script_cast_members.play_script_id = play_scripts.id) = sqlc.arg(participant_count)",
					),
					buildCursorLimitClause(playScriptsTableName),
				)),
			},
		}
	default:
		return nil
	}
}
