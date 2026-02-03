package integration

const (
	exampleQuantity = 5
)

func defaultIgnoredFields(additionalFields ...string) []string {
	return append([]string{"CreatedAt", "LastUpdatedAt", "ArchivedAt"}, additionalFields...)
}
