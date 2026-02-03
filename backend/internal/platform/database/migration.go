package database

// Migration represents a database migration.
type Migration struct {
	Description string
	Query       string
}
