package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	playScriptLinesTableName = "play_script_lines"
	castMemberIDColumn       = "cast_member_id"
	lineTextColumn           = "line_text"
	lineIndexColumn          = "line_index"
	directionColumn          = "direction"
	scrollModeColumn         = "scroll_mode"
	autoScrollWpmColumn      = "auto_scroll_wpm"
)

func init() {
	registerTableName(playScriptLinesTableName)
}

var playScriptLinesColumns = []string{
	idColumn,
	playScriptIDColumn,
	castMemberIDColumn,
	lineTextColumn,
	lineIndexColumn,
	directionColumn,
	scrollModeColumn,
	autoScrollWpmColumn,
	createdAtColumn,
}

func buildPlayScriptLinesQueries(database string) []*Query {
	switch database {
	case postgres:
		insertColumns := filterForInsert(playScriptLinesColumns)
		fullSelectColumns := applyToEach(playScriptLinesColumns, func(_ int, s string) string {
			return fullColumnName(playScriptLinesTableName, s)
		})

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreatePlayScriptLine",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					playScriptLinesTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetPlayScriptLine",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					playScriptLinesTableName,
					playScriptLinesTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetPlayScriptLinesForPlayScript",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
ORDER BY %s.%s ASC;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					playScriptLinesTableName,
					playScriptLinesTableName, playScriptIDColumn, playScriptIDColumn,
					playScriptLinesTableName, lineIndexColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdatePlayScriptLine",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s)
WHERE %s = sqlc.arg(%s);`,
					playScriptLinesTableName,
					castMemberIDColumn, castMemberIDColumn,
					lineTextColumn, lineTextColumn,
					lineIndexColumn, lineIndexColumn,
					directionColumn, directionColumn,
					scrollModeColumn, scrollModeColumn,
					autoScrollWpmColumn, autoScrollWpmColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "DeletePlayScriptLine",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s
WHERE %s = sqlc.arg(%s);`,
					playScriptLinesTableName,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "DeletePlayScriptLinesForPlayScript",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s
WHERE %s = sqlc.arg(%s);`,
					playScriptLinesTableName,
					playScriptIDColumn, playScriptIDColumn,
				)),
			},
		}
	default:
		return nil
	}
}
