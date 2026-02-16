package playscripts

// Repository describes persistence capabilities for play scripts, script tags, and ratings.
type Repository interface {
	PlayScriptDataManager
	PlayScriptRatingDataManager
	ScriptTagDataManager
}
