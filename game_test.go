package main

import "testing"

func TestParseManifest(t *testing.T) {
	var data = []byte(`
---
manifest:
    name: TestName
`)
	d := Data{}
	err := d.parseYaml([]byte(data))
	if err != nil {
		t.Error("Failed to parse manifest")
	}

	if d.Manifest.Name != "TestName" {
		t.Error("Failed to parse name")
	}
}

func TestParseSettings(t *testing.T) {
	var data = []byte(`
---
settings:
    width: 888
    height: 555
    fullscreen: true
`)
	d := Data{}
	err := d.parseYaml([]byte(data))
	if err != nil {
		t.Error("Failed to parse options")
	}

	if d.Settings.Width != 888 {
		t.Error("Failed to parse width")
	}

	if d.Settings.Height != 555 {
		t.Error("Failed to parse height")
	}

	if d.Settings.Fullscreen != true {
		t.Error("Failed to parse fullscreen")
	}
}
