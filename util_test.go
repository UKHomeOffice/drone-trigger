package main

import "testing"

func TestParsePairs(t *testing.T) {
	s := []string{"FOO=bar/something:tag", "BAR=", "INVALID"}
	p := parsePairs(s)
	if p["FOO"] != "bar/something:tag" {
		t.Errorf("Wanted %q, got %q.", "bar/something:tag", p["FOO"])
	}
	if _, exists := p["BAR"]; !exists {
		t.Error("Missing a key with no value. Keys with empty values are also valid.")
	}
	if _, exists := p["INVALID"]; exists {
		t.Error("Keys without an equal sign suffix are invalid.")
	}
}
