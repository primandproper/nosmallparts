package bootstrap

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
)

//go:embed shakespeare/*.json
var shakespeareFS embed.FS

// ShakespearePlayScripts returns all play script creation requests from the embedded
// shakespeare/ JSON files, sorted by filename.
func ShakespearePlayScripts() ([]*playscripts.PlayScriptCreationRequestInput, error) {
	entries, err := shakespeareFS.ReadDir("shakespeare")
	if err != nil {
		return nil, fmt.Errorf("read shakespeare dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	result := make([]*playscripts.PlayScriptCreationRequestInput, 0, len(names))
	for _, name := range names {
		path := filepath.Join("shakespeare", name)
		data, err := shakespeareFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var input playscripts.PlayScriptCreationRequestInput
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %w", path, err)
		}
		result = append(result, &input)
	}
	return result, nil
}
