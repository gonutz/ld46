// author: Jacky Boen

package main

import (
	"fmt"
	"github.com/gonutz/go-sdl2/img"
	"github.com/gonutz/go-sdl2/sdl"
	"os"
)

var winTitle string = "Go-SDL2 Texture"
var winWidth, winHeight int32 = 800, 600
var imageName string = "../../assets/test.png"

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var texture *sdl.Texture
	var src, dst sdl.Rect
	var err error

	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer renderer.Destroy()

	image, err := img.Load(imageName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load PNG: %s\n", err)
		return 3
	}
	defer image.Free()

	texture, err = renderer.CreateTextureFromSurface(image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
		return 4
	}
	defer texture.Destroy()

	src = sdl.Rect{0, 0, 512, 512}
	dst = sdl.Rect{100, 50, 512, 512}

	renderer.Clear()
	renderer.SetDrawColor(255, 0, 0, 255)
	renderer.FillRect(&sdl.Rect{0, 0, int32(winWidth), int32(winHeight)})
	renderer.Copy(texture, &src, &dst)
	renderer.Present()

	sdl.Delay(2000)

	return 0
}

func main() {
	os.Exit(run())
}
