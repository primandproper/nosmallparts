package manager

import (
	"slices"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"

	"github.com/stretchr/testify/mock"
)

func eventMatches(eventType string, keys []string) any {
	return mock.MatchedBy(func(message *audit.DataChangeMessage) bool {
		allContextKeys := []string{}
		for k := range message.Context {
			allContextKeys = append(allContextKeys, k)
		}

		slices.Sort(keys)
		slices.Sort(allContextKeys)
		allKeysMatch := slices.Equal(keys, allContextKeys)
		eventTypeMatches := message.EventType == eventType
		result := allKeysMatch && eventTypeMatches

		return result
	})
}
