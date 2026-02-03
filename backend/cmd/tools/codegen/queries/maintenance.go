package main

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

var (
	allTablesHat sync.Mutex
	allTables    = map[string]bool{}
)

func registerTableName(table string) {
	allTablesHat.Lock()
	defer allTablesHat.Unlock()
	allTables[table] = true
}

func getAllTables() []string {
	allTablesHat.Lock()
	defer allTablesHat.Unlock()

	tables := make([]string, 0, len(allTables))
	for t := range allTables {
		tables = append(tables, t)
	}

	slices.Sort(tables)

	return tables
}

func buildMaintenanceQueries(database string) []*Query {
	switch database {
	case postgres:
		const oneDayAgo = `(NOW() - interval '1 day')`
		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "DeleteExpiredOAuth2ClientTokens",
					Type: ExecRowsType,
				},
				Content: fmt.Sprintf(`DELETE FROM %s WHERE %s < %s AND %s < %s AND %s < %s;`, oauth2ClientTokensTableName, codeExpiresAtColumn, oneDayAgo, accessExpiresAtColumn, oneDayAgo, refreshExpiresAtColumn, oneDayAgo),
			},
			{
				Annotation: QueryAnnotation{
					Name: "DestroyAllData",
					Type: ExecType,
				},
				Content: fmt.Sprintf(`TRUNCATE %s CASCADE;`, strings.Join(getAllTables(), ", ")),
			},
		}
	default:
		return nil
	}
}
