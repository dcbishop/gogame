package main

import "testing"

func TestYAML(t *testing.T) {
	var data = `
name: TestName
width: 888
height: 555
fullscreen: false
`
	opts, err := parseOptions([]byte(data))
	if err != nil {
		t.Error("Failed to parse options")
	}

	if opts.Name != "TestName" {
		t.Error("Failed to parse name")
	}
}
