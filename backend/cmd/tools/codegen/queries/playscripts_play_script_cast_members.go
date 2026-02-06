package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	playScriptCastMembersTableName = "play_script_cast_members"
	playScriptIDColumn             = "play_script_id"
	sortOrderColumn                = "sort_order"
)

func init() {
	registerTableName(playScriptCastMembersTableName)
}

var playScriptCastMembersColumns = []string{
	idColumn,
	playScriptIDColumn,
	nameColumn,
	sortOrderColumn,
	createdAtColumn,
}

func buildPlayScriptCastMembersQueries(database string) []*Query {
	switch database {
	case postgres:
		insertColumns := filterForInsert(playScriptCastMembersColumns)
		fullSelectColumns := applyToEach(playScriptCastMembersColumns, func(_ int, s string) string {
			return fullColumnName(playScriptCastMembersTableName, s)
		})

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreatePlayScriptCastMember",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					playScriptCastMembersTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetPlayScriptCastMember",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					playScriptCastMembersTableName,
					playScriptCastMembersTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetPlayScriptCastMembersForPlayScript",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
ORDER BY %s.%s ASC;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					playScriptCastMembersTableName,
					playScriptCastMembersTableName, playScriptIDColumn, playScriptIDColumn,
					playScriptCastMembersTableName, sortOrderColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdatePlayScriptCastMember",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s)
WHERE %s = sqlc.arg(%s);`,
					playScriptCastMembersTableName,
					nameColumn, nameColumn,
					sortOrderColumn, sortOrderColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "DeletePlayScriptCastMember",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s
WHERE %s = sqlc.arg(%s);`,
					playScriptCastMembersTableName,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "DeletePlayScriptCastMembersForPlayScript",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s
WHERE %s = sqlc.arg(%s);`,
					playScriptCastMembersTableName,
					playScriptIDColumn, playScriptIDColumn,
				)),
			},
		}
	default:
		return nil
	}
}
