package bootstrap

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
)

const (
	jsonExtension = ".json"
)

//go:embed shakespeare/*.json movies/*.json tvshows/*.json
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
		if !e.IsDir() && filepath.Ext(e.Name()) == jsonExtension {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	result := make([]*playscripts.PlayScriptCreationRequestInput, 0, len(names))
	for _, name := range names {
		path := filepath.Join("shakespeare", name)
		var data []byte
		data, err = shakespeareFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var input playscripts.PlayScriptCreationRequestInput
		if err = json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %w", path, err)
		}
		result = append(result, &input)
	}
	return result, nil
}

// IdealLineCountMin is the minimum number of lines for an "ideal" play script.
// Scripts with fewer lines (e.g. prologues) are too short for a meaningful session.
const IdealLineCountMin = 30

// IdealLineCountMax is the maximum number of lines for an "ideal" play script.
// Scripts with more lines typically exceed ~15 minutes of performance time.
const IdealLineCountMax = 120

// MoviePlayScripts returns all play script creation requests from the embedded
// movies/ JSON files, sorted by filename.
func MoviePlayScripts() ([]*playscripts.PlayScriptCreationRequestInput, error) {
	entries, err := shakespeareFS.ReadDir("movies")
	if err != nil {
		return nil, fmt.Errorf("read movies dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == jsonExtension {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	result := make([]*playscripts.PlayScriptCreationRequestInput, 0, len(names))
	for _, name := range names {
		path := filepath.Join("movies", name)
		var data []byte
		data, err = shakespeareFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var input playscripts.PlayScriptCreationRequestInput
		if err = json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %w", path, err)
		}
		result = append(result, &input)
	}
	return result, nil
}

// TVShowPlayScripts returns all play script creation requests from the embedded
// tvshows/ JSON files, sorted by filename.
func TVShowPlayScripts() ([]*playscripts.PlayScriptCreationRequestInput, error) {
	entries, err := shakespeareFS.ReadDir("tvshows")
	if err != nil {
		return nil, fmt.Errorf("read tvshows dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == jsonExtension {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	result := make([]*playscripts.PlayScriptCreationRequestInput, 0, len(names))
	for _, name := range names {
		path := filepath.Join("tvshows", name)
		var data []byte
		data, err = shakespeareFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var input playscripts.PlayScriptCreationRequestInput
		if err = json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %w", path, err)
		}
		result = append(result, &input)
	}
	return result, nil
}

// IdealPlayScripts returns play scripts that are appropriately sized for a ~15 minute
// session: not too short (e.g. prologues) and not too long. It filters the Shakespeare
// collection to only include scripts with line counts between IdealLineCountMin and
// IdealLineCountMax (inclusive).
func IdealPlayScripts() ([]*playscripts.PlayScriptCreationRequestInput, error) {
	all, err := ShakespearePlayScripts()
	if err != nil {
		return nil, err
	}
	result := make([]*playscripts.PlayScriptCreationRequestInput, 0, len(all))
	for _, script := range all {
		lineCount := len(script.Lines)
		if lineCount >= IdealLineCountMin && lineCount <= IdealLineCountMax {
			result = append(result, script)
		}
	}
	return result, nil
}
