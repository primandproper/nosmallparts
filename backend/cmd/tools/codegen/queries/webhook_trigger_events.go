package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	webhookTriggerEventsTableName = "webhook_trigger_events"
	belongsToWebhookColumn        = "belongs_to_webhook"
	triggerEventColumn            = "trigger_event"
)

func init() {
	registerTableName(webhookTriggerEventsTableName)
}

var (
	webhookTriggerEventsColumns = []string{
		idColumn,
		triggerEventColumn,
		belongsToWebhookColumn,
		createdAtColumn,
		archivedAtColumn,
	}
)

func buildWebhookTriggerEventsQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(webhookTriggerEventsColumns)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreateWebhookTriggerEvent",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					webhookTriggerEventsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveWebhookTriggerEvent",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					webhookTriggerEventsTableName,
					archivedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
					belongsToWebhookColumn, belongsToWebhookColumn,
				)),
			},
		}
	default:
		return nil
	}
}
