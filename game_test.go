package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
	game.data = failsafeData()
	game.data.Manifest.Name = "OldName"

	data := Data{}
	data.Manifest.Name = "NewName"
	data.Settings.Width = 1024
	data.Settings.Height = 768

	game.ApplyDataChanges(&data)

	if game.data.Manifest.Name != "NewName" {
		t.Error("Failed to apply new name to games data:", game.data.Manifest.Name, data.Manifest.Name)
	}

	if game.data.Settings.Width != 1024 {
		t.Error("Failed to apply new width to games data")
	}

	if game.data.Settings.Height != 768 {
		t.Error("Failed to apply new height to games data")
	}
}

func TestApplyDataChangesIgnoresMissingChanges(t *testing.T) {
	game := Game{}
	game.data.Manifest.Name = "OldName"
	game.data.Settings.Width = 1024
	game.data.Settings.Height = 768

	data := magicData()

	game.ApplyDataChanges(&data)

	if game.data.Manifest.Name != "OldName" {
		t.Error("Applied empty name to games data")
	}

	if game.data.Settings.Width != 1024 {
		t.Error("Applied magic width to games data", game.data.Settings.Width)
	}

	if game.data.Settings.Height != 768 {
		t.Error("Applied magic height to games data", game.data.Settings.Height)
	}
}

func TestFileWatcher(t *testing.T) {
	data := []byte(`
---
manifest:
    name: InitialName
`)

	Convey("Construct new Game", t, func() {
		game := NewGame()
		So(game.data.Manifest.Name, ShouldEqual, failsafeGameName)
		var dir string
		var err error
		Convey("Create data directory", func() {
			dir, err = ioutil.TempDir(".", "test")
			So(err, ShouldBeNil)

			Convey("Add YAML file", func() {
				path := path.Join(dir, "data.yaml")
				ioutil.WriteFile(path, data, 0600)

				Convey("Set game's data directory", func() {
					game.SetDataDirectory(dir)
					So(game.watcher, ShouldNotBeNil)

					Convey("It should load the name.", func() {
						go game.consumeAllFileEvents()
						game.touched <- path
						close(game.quit)
						game.Run()

						So(game.data.Manifest.Name, ShouldEqual, "InitialName")

						os.RemoveAll(dir)
					})
				})
			})
		})
	})
}
