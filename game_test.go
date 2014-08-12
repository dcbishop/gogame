package main

import "testing"

func TestParseManifest(t *testing.T) {
	var data = `
name: TestName
`
	manifest, err := parseManifest([]byte(data))
	if err != nil {
		t.Error("Failed to parse options")
	}

	if manifest.Name != "TestName" {
		t.Error("Failed to parse name")
	}
}

func TestParseSettings(t *testing.T) {
	var data = `
width: 888
height: 555
fullscreen: true
`
	settings, err := parseSettings([]byte(data))
	if err != nil {
		t.Error("Failed to parse options")
	}

	if settings.Width != 888 {
		t.Error("Failed to parse width")
	}

	if settings.Height != 555 {
		t.Error("Failed to parse height")
	}

	if settings.Fullscreen != true {
		t.Error("Failed to parse fullscreen")
	}
}
