package main

import (
	"errors"

	"github.com/veandco/go-sdl2/sdl"
)

// Window is where we render to.
type Window interface {
	Run()
}

// SDLWindow is a Window using SDL2
type SDLWindow struct {
	window  *sdl.Window
	context sdl.GLContext
}

// NewSDLWindow constructs a SDLWindow.
func NewSDLWindow(name string, width int, height int) (*SDLWindow, error) {
	window := new(SDLWindow)

	window.window = sdl.CreateWindow(name, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_OPENGL)
	window.context = sdl.GL_CreateContext(window.window)

	if window.context == nil {
		return nil, errors.New("Could not create OpenGL context.")
	}
	return window, nil
}

func (window SDLWindow) Run() {
	surface := window.window.GetSurface()

	xpos := int32(0)

	for {
		xpos = (xpos + 1) % 300
		rect := sdl.Rect{0 + xpos, 0, 200 + xpos, 200}
		surface.FillRect(&rect, 0xffff0000)
		window.window.UpdateSurface()
	}

	sdl.Delay(5000)
	window.window.Destroy()
}
