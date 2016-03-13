package main

import (
	"errors"
	"log"

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
	render  *sdl.Renderer
	debug   struct {
		xpos int32
		rect sdl.Rect
	}
	initialWidth  int
	initialHeight int
}

// NewSDLWindow constructs a SDLWindow.
func NewSDLWindow() (*SDLWindow, error) {
	return newSDLWindowSettings(failsafeGameName, 0, 0)
}

//newSDLWindowSettings constructs a SDLWindow with initial settings
func newSDLWindowSettings(name string, width int, height int) (*SDLWindow, error) {
	sdl.Init(sdl.INIT_EVERYTHING)
	window := new(SDLWindow)

	w, err := sdl.CreateWindow(name, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_OPENGL)
	if err != nil {
		return nil, err
	}
	window.window = w
	window.window.Show()
	window.context, err = sdl.GL_CreateContext(window.window)
	if err != nil {
		return nil, err
	}

	window.surface, err = window.window.GetSurface()
	if err != nil {
		return nil, err
	}

	r, err := window.window.GetRenderer()
	window.render = r
	if err != nil {
		log.Println(err)
	}

	if window.context == nil {
		return nil, errors.New("Could not create OpenGL context.")
	}
	window.debug.xpos = 100
	return window, nil
}

// Update redraws the Window.
func (window *SDLWindow) Update() {
	window.updateDebugSquarePosition()
	window.draw()
}

// Updates the position of the red debug square.
func (window *SDLWindow) updateDebugSquarePosition() {
	w, _ := window.window.GetSize()
	window.debug.xpos = (window.debug.xpos + 1) % int32(w)
	window.debug.rect = sdl.Rect{
		X: 0 + window.debug.xpos,
		Y: 0,
		W: 20,
		H: 20,
	}
}

// Renders the window.
func (window *SDLWindow) draw() {
	window.clear()
	drawDebugRedSquare(window.render, &window.debug.rect)
	window.render.Present()
}

// Clears the window to white.
func (window *SDLWindow) clear() {
	window.render.SetDrawColor(255, 255, 255, 255)
	err := window.render.Clear()
	if err != nil {
		log.Println(err)
	}
}

func drawDebugRedSquare(render *sdl.Renderer, rect *sdl.Rect) {
	render.SetDrawColor(255, 0, 0, 255)
	render.FillRect(rect)
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
		window.surface, _ = window.window.GetSurface()
	}
}
