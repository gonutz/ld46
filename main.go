package main

import "github.com/gonutz/prototype/draw"

var (
	windowTitle      = "LD46"
	windowFullscreen = true

	// I want to try out a blue color palette for this game.
	blue0 = rgb(255, 255, 255)
	blue1 = rgb(200, 240, 255)
	blue2 = rgb(140, 220, 255)
	blue3 = rgb(65, 200, 255)
	blue4 = rgb(0, 175, 255)
	blue5 = rgb(0, 155, 235)
	blue6 = rgb(0, 140, 210)
	blue7 = rgb(0, 120, 180)
	blue8 = rgb(0, 105, 155)
	blues = []draw.Color{
		blue0, blue1, blue2, blue3, blue4, blue5, blue6, blue7, blue8,
	}
)

func main() {
	err := draw.RunWindow(windowTitle, 800, 600, func(window draw.Window) {
		if window.WasKeyPressed(draw.KeyEscape) {
			window.Close()
		}
		if window.WasKeyPressed(draw.KeyEnter) &&
			(window.IsKeyDown(draw.KeyLeftAlt) || window.IsKeyDown(draw.KeyRightAlt)) {
			windowFullscreen = !windowFullscreen
		}
		window.SetFullscreen(windowFullscreen)
		windowW, windowH := window.Size()

		for i, c := range blues {
			window.FillRect(0, i*windowH/len(blues), windowW, windowH/len(blues)+1, c)
		}
	})
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func rgb(r, g, b uint8) draw.Color {
	const f = 1.0 / 255.0
	return draw.RGB(float32(r)*f, float32(g)*f, float32(b)*f)
}
