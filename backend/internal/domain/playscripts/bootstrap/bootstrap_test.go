package bootstrap

import (
	"testing"
)

func TestShakespearePlayScripts(t *testing.T) {
	scripts, err := ShakespearePlayScripts()
	if err != nil {
		t.Fatalf("ShakespearePlayScripts: %v", err)
	}
	if len(scripts) == 0 {
		t.Fatal("expected at least one script")
	}
	// Spot check first script
	first := scripts[0]
	if first.Title == "" {
		t.Error("expected non-empty title")
	}
	if len(first.CastMembers) == 0 {
		t.Error("expected at least one cast member")
	}
	if len(first.Lines) == 0 {
		t.Error("expected at least one line")
	}
}

func TestIdealPlayScripts(t *testing.T) {
	ideal, err := IdealPlayScripts()
	if err != nil {
		t.Fatalf("IdealPlayScripts: %v", err)
	}
	all, err := ShakespearePlayScripts()
	if err != nil {
		t.Fatalf("ShakespearePlayScripts: %v", err)
	}
	if len(ideal) > len(all) {
		t.Errorf("IdealPlayScripts returned more scripts (%d) than ShakespearePlayScripts (%d)", len(ideal), len(all))
	}
	for i, script := range ideal {
		lineCount := len(script.Lines)
		if lineCount < IdealLineCountMin {
			t.Errorf("script %d %q has %d lines, below minimum %d", i, script.Title, lineCount, IdealLineCountMin)
		}
		if lineCount > IdealLineCountMax {
			t.Errorf("script %d %q has %d lines, above maximum %d", i, script.Title, lineCount, IdealLineCountMax)
		}
	}
	// Spot check: Comedy of Errors Act 1 (47 lines) should be included
	var foundComedy bool
	for _, s := range ideal {
		if s.Title == "The Comedy of Errors (Act 1)" {
			foundComedy = true
			break
		}
	}
	if !foundComedy {
		t.Error("expected Comedy of Errors Act 1 (47 lines) to be in IdealPlayScripts")
	}
}
