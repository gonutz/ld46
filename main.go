package main

import "github.com/gonutz/prototype/draw"

var (
	windowTitle      = "LD46"
	windowFullscreen = true

	tileSize = 64

	handClosing = false
	handFrame   = newFrameTimer(3).clamp(0, len(handCursors)-1)
	handCursors = []string{
		"open_hand_cursor.png",
		"closing_hand_cursor_1.png",
		"closing_hand_cursor_2.png",
	}

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

		for y := 0; y < windowH; y++ {
			window.DrawLine(
				0, y, windowW, y,
				lerpColor(blue1, blue2, float32(y)/float32(windowH-1)),
			)
		}
		for x := 0; x < 100; x++ {
			window.DrawImageFile("solid_tile.png", x*tileSize, windowH-tileSize)
		}
		for x := 0; x < 2; x++ {
			for y := 0; y < 100; y++ {
				window.DrawImageFile("solid_tile.png", x*tileSize, windowH-(y*tileSize))
			}
		}
		window.DrawImageFile("door_tile.png", tileSize, windowH-3*tileSize)
		window.DrawImageFile("draggable_tile.png", 5*tileSize, windowH-5*tileSize)
		window.DrawImageFile("draggable_tile.png", 6*tileSize, windowH-5*tileSize)
		window.DrawImageFile("draggable_tile.png", 7*tileSize, windowH-5*tileSize)

		mx, my := window.MousePosition()
		if window.IsMouseDown(draw.LeftButton) {
			handFrame.inc()
		} else {
			handFrame.dec()
		}
		cursorFrame := clamp(handFrame.value(), 0, len(handCursors)-1)
		window.DrawImageFile(handCursors[cursorFrame], mx-20, my-20)
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

func lerpColor(a, b draw.Color, t float32) draw.Color {
	return draw.Color{
		R: a.R*t + b.R*(1.0-t),
		G: a.G*t + b.G*(1.0-t),
		B: a.B*t + b.B*(1.0-t),
		A: a.A*t + b.A*(1.0-t),
	}
}

func clamp(a, min, max int) int {
	if a < min {
		a = min
	}
	if a > max {
		a = max
	}
	return a
}

func newFrameTimer(div int) *frameTimer {
	return &frameTimer{
		divider:      div,
		undividedMin: -9999999,
		undividedMax: 9999999,
	}
}

type frameTimer struct {
	divider      int
	undivided    int
	undividedMin int
	undividedMax int
}

func (t *frameTimer) clamp(min, max int) *frameTimer {
	t.undividedMin = min * t.divider
	t.undividedMax = max * t.divider
	return t
}

func (t *frameTimer) inc() {
	if t.undivided < t.undividedMax {
		t.undivided++
	}
}

func (t *frameTimer) dec() {
	if t.undivided > t.undividedMin {
		t.undivided--
	}
}

func (t *frameTimer) value() int {
	return t.undivided / t.divider
}
