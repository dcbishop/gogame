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
    windowmode: fullscreen
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

	if d.Settings.WindowMode != windowModeFullscreen {
		t.Error("Failed to parse fullscreen:", d.Settings.WindowMode)
	}
}

func TestApplyDataChanges(t *testing.T) {
	game := Game{}
	game.data.Manifest.Name = "OldName"

	data := Data{}
	data.Manifest.Name = "NewName"

	game.ApplyDataChanges(&data)

	if game.data.Manifest.Name != "NewName" {
		t.Error("Failed to apply new name to games data:", game.data.Manifest.Name, data.Manifest.Name)
	}
}

func TestApplyDataChangesIgnoresMissingChanges(t *testing.T) {
	game := Game{}
	game.data.Manifest.Name = "OldName"

	data := magicData()

	game.ApplyDataChanges(&data)

	if game.data.Manifest.Name != "OldName" {
		t.Error("Applied empty name to games data")
	}
}
