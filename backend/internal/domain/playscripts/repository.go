package playscripts

// Repository describes persistence capabilities for play scripts and script tags.
type Repository interface {
	PlayScriptDataManager
	ScriptTagDataManager
}
