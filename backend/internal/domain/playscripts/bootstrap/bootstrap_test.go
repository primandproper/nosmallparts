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
