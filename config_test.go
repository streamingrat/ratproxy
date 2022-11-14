package main

import "testing"

func TestNewConfig(t *testing.T) {

	c := NewConfig()
	if c.Listen != "0.0.0.0:1414" {
		t.Fatalf("Did not get expected listen")
	}

	if len(c.Targets) != 2 {
		t.Fatalf("Did not get 2 expected targets, got %v\n", len(c.Targets))
	}
}
