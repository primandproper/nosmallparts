package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	scriptTagsTableName     = "script_tags"
	playScriptTagsTableName = "play_script_tags"
	scriptTagIDColumn       = "script_tag_id"
)

func init() {
	registerTableName(scriptTagsTableName)
	registerTableName(playScriptTagsTableName)
}

var scriptTagsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	createdAtColumn,
}

var playScriptTagsColumns = []string{
	playScriptIDColumn,
	scriptTagIDColumn,
	createdAtColumn,
	archivedAtColumn,
}

func buildScriptTagsQueries(database string) []*Query {
	switch database {
	case postgres:
		scriptTagsInsertColumns := filterForInsert(scriptTagsColumns)
		scriptTagsFullSelect := applyToEach(scriptTagsColumns, func(_ int, s string) string {
			return fullColumnName(scriptTagsTableName, s)
		})

		return []*Query{
			// script_tags CRUD
			{
				Annotation: QueryAnnotation{Name: "CreateScriptTag", Type: ExecType},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					scriptTagsTableName,
					strings.Join(scriptTagsInsertColumns, ",\n\t"),
					strings.Join(applyToEach(scriptTagsInsertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{Name: "GetScriptTag", Type: OneType},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s);`,
					strings.Join(scriptTagsFullSelect, ",\n\t"),
					scriptTagsTableName,
					scriptTagsTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{Name: "GetScriptTagByName", Type: OneType},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s);`,
					strings.Join(scriptTagsFullSelect, ",\n\t"),
					scriptTagsTableName,
					scriptTagsTableName, nameColumn, nameColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{Name: "ListScriptTags", Type: ManyType},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
ORDER BY %s.%s ASC;`,
					strings.Join(scriptTagsFullSelect, ",\n\t"),
					scriptTagsTableName,
					scriptTagsTableName, nameColumn,
				)),
			},
			// play_script_tags: AddTagToPlayScript - upsert to handle re-adding archived tag
			{
				Annotation: QueryAnnotation{Name: "AddTagToPlayScript", Type: ExecType},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s,
	%s,
	%s
) VALUES (
	sqlc.arg(%s),
	sqlc.arg(%s),
	%s
)
ON CONFLICT (%s, %s) DO UPDATE SET
	%s = NULL;`,
					playScriptTagsTableName,
					playScriptIDColumn,
					scriptTagIDColumn,
					createdAtColumn,
					playScriptIDColumn,
					scriptTagIDColumn,
					currentTimeExpression,
					playScriptIDColumn, scriptTagIDColumn,
					archivedAtColumn,
				)),
			},
			// play_script_tags: RemoveTagFromPlayScript (soft delete)
			{
				Annotation: QueryAnnotation{Name: "RemoveTagFromPlayScript", Type: ExecRowsType},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					playScriptTagsTableName,
					archivedAtColumn, currentTimeExpression,
					archivedAtColumn,
					playScriptIDColumn, playScriptIDColumn,
					scriptTagIDColumn, scriptTagIDColumn,
				)),
			},
			// GetTagsForPlayScript - join script_tags with play_script_tags where active
			{
				Annotation: QueryAnnotation{Name: "GetTagsForPlayScript", Type: ManyType},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s.%s,
	%s.%s,
	%s.%s,
	%s.%s
FROM %s
JOIN %s ON %s.%s = %s.%s
	AND %s.%s IS NULL
WHERE %s.%s = sqlc.arg(%s);`,
					scriptTagsTableName, idColumn,
					scriptTagsTableName, nameColumn,
					scriptTagsTableName, descriptionColumn,
					scriptTagsTableName, createdAtColumn,
					scriptTagsTableName,
					playScriptTagsTableName,
					playScriptTagsTableName, scriptTagIDColumn, scriptTagsTableName, idColumn,
					playScriptTagsTableName, archivedAtColumn,
					playScriptTagsTableName, playScriptIDColumn, playScriptIDColumn,
				)),
			},
			// GetPlayScriptsWithTag - scripts that have a tag with given name, with pagination
			{
				Annotation: QueryAnnotation{Name: "GetPlayScriptsWithTag", Type: ManyType},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	play_scripts.id,
	play_scripts.title,
	play_scripts.description,
	play_scripts.author,
	play_scripts.is_public,
	play_scripts.created_at,
	play_scripts.last_updated_at,
	play_scripts.archived_at,
	play_scripts.created_by_user,
	play_scripts.belongs_to_account,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		JOIN %s ON %s.%s = play_scripts.id AND %s.%s IS NULL
		JOIN %s ON %s.%s = %s.%s AND %s.%s = sqlc.arg(%s)
		WHERE play_scripts.archived_at IS NULL
			AND play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	) AS filtered_count,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		JOIN %s ON %s.%s = play_scripts.id AND %s.%s IS NULL
		JOIN %s ON %s.%s = %s.%s AND %s.%s = sqlc.arg(%s)
		WHERE play_scripts.archived_at IS NULL
	) AS total_count
FROM play_scripts
JOIN %s ON %s.%s = play_scripts.id AND %s.%s IS NULL
JOIN %s ON %s.%s = %s.%s AND %s.%s = sqlc.arg(%s)
WHERE play_scripts.archived_at IS NULL
	AND play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND play_scripts.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY play_scripts.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);`,
					playScriptTagsTableName, playScriptTagsTableName, playScriptIDColumn, playScriptTagsTableName, archivedAtColumn,
					scriptTagsTableName, scriptTagsTableName, idColumn, playScriptTagsTableName, scriptTagIDColumn, scriptTagsTableName, nameColumn, nameColumn,
					playScriptTagsTableName, playScriptTagsTableName, playScriptIDColumn, playScriptTagsTableName, archivedAtColumn,
					scriptTagsTableName, scriptTagsTableName, idColumn, playScriptTagsTableName, scriptTagIDColumn, scriptTagsTableName, nameColumn, nameColumn,
					playScriptTagsTableName, playScriptTagsTableName, playScriptIDColumn, playScriptTagsTableName, archivedAtColumn,
					scriptTagsTableName, scriptTagsTableName, idColumn, playScriptTagsTableName, scriptTagIDColumn, scriptTagsTableName, nameColumn, nameColumn,
				)),
			},
		}
	default:
		return nil
	}
}
