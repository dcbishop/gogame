package main

import (
	"errors"

	"github.com/veandco/go-sdl2/sdl"
)

// Window is where we render to.
type Window interface {
	Update()
	SetTitle(name string)
	SetSize(width int, height int)
	Destroy()
}

// SDLWindow is a Window using SDL2
type SDLWindow struct {
	title   string
	window  *sdl.Window
	context sdl.GLContext
	surface *sdl.Surface
	debug   struct {
		xpos int32
		rect sdl.Rect
	}
	initialWidth  int
	initialHeight int
}

// NewSDLWindow constructs a SDLWindow.
func NewSDLWindow(name string, width int, height int) (*SDLWindow, error) {
	sdl.Init(sdl.INIT_EVERYTHING)
	window := new(SDLWindow)

	window.window = sdl.CreateWindow(name, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_OPENGL)
	window.window.Show()
	window.context = sdl.GL_CreateContext(window.window)
	window.surface = window.window.GetSurface()

	if window.context == nil {
		return nil, errors.New("Could not create OpenGL context.")
	}
	window.debug.xpos = 100
	return window, nil
}

// Update redraws the Window
func (window *SDLWindow) Update() {
	window.debug.xpos = (window.debug.xpos + 1)
	window.debug.rect = sdl.Rect{0 + window.debug.xpos, 0, 200 + window.debug.xpos, 200}

	window.surface.FillRect(&window.debug.rect, 0xffff0000)
	window.window.UpdateSurface()
}

// Destroy cleans up the Window
func (window *SDLWindow) Destroy() {
	window.window.Destroy()
}

// SetTitle sets the title of the Window
func (window *SDLWindow) SetTitle(name string) {
	if window.title != name {
		window.title = name
		window.window.SetTitle(name)
	}
}

// SetSize sets the size of the Window
func (window *SDLWindow) SetSize(width int, height int) {
	if width != window.initialWidth || height != window.initialHeight {
		window.initialWidth = width
		window.initialHeight = height
		window.window.SetSize(width, height)
		window.surface = window.window.GetSurface()
	}
}
