package main

import (
	"strings"

	"github.com/gonutz/prototype/draw"
)

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

	leftMouseWasDown     = false
	movingTile           *tile
	previewDx, previewDy int

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

	tileSolid     = "solid_tile.png"       // x
	tileDrag      = "draggable_tile.png"   // o
	tileDoor      = "door_tile.png"        // D
	tileHighlight = "highlighted_tile.png" // only at runtime
	tileMoving    = "highlighted_tile.png" // only at runtime
	tilePreview   = "preview_tile.png"     // only at runtime

	level1 = parseLevel(`
	xxxxxxxxxxxxxxxxxxxxx
	x        xxx        x
	x                   x
	x                   x
	x                   x
	x                   x
	x                   x
	x                   x
	x        ooo        x
	x                   x
	x                   D
	x                    
	xxxxxxxxx   xxxxxxxxx
	`)
)

func main() {
	level := level1
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

		// Handle input.
		mx, my := window.MousePosition()
		leftMouseDown := window.IsMouseDown(draw.LeftButton)
		// Animate the hand opening/closing.
		if leftMouseDown {
			handFrame.inc()
		} else {
			handFrame.dec()
		}
		if !leftMouseWasDown && leftMouseDown {
			for i, t := range level.tiles {
				if t.image == tileDrag && t.contains(mx, my) {
					movingTile = &level.tiles[i]
					movingTile.image = tileMoving
					previewDx = mx - t.x*tileSize
					previewDy = my - t.y*tileSize
				}
			}
		}
		if !leftMouseDown && movingTile != nil {
			movingTile.image = tileDrag
			movingTile = nil
		}
		if movingTile != nil {
			newX := (mx - previewDx + tileSize/2) / tileSize
			newY := (my - previewDy + tileSize/2) / tileSize
			t := level.tileAt(newX, newY)
			if t == nil || t == movingTile {
				movingTile.x = newX
				movingTile.y = newY
			}
		}

		// Draw background sky as light blue gradient.
		for y := 0; y < windowH; y++ {
			window.DrawLine(
				0, y, windowW, y,
				lerpColor(blue1, blue2, float32(y)/float32(windowH-1)),
			)
		}

		// Draw tiles.
		for _, t := range level.tiles {
			image := t.image
			if movingTile == nil && t.image == tileDrag && t.contains(mx, my) {
				image = tileHighlight
			}
			window.DrawImageFile(image, t.x*tileSize, t.y*tileSize)
		}

		// Draw mouse cursor.
		window.DrawImageFile(handCursors[handFrame.value()], mx-20, my-20)
		if movingTile != nil {
			window.DrawImageFile(tilePreview, mx-previewDx, my-previewDy)
		}

		leftMouseWasDown = leftMouseDown
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

func parseLevel(s string) *level {
	mapping := map[rune]string{
		'x': tileSolid,
		'o': tileDrag,
		'D': tileDoor,
	}
	var tiles []tile
	y := 0
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for x, r := range line {
			x := x
			if image, ok := mapping[r]; ok {
				tiles = append(tiles, tile{x: x, y: y, image: image})
			}
		}
		y++
	}
	return &level{tiles: tiles}
}

type level struct {
	tiles []tile
}

func (l *level) tileAt(x, y int) *tile {
	for i, t := range l.tiles {
		if x == t.x && y == t.y {
			return &l.tiles[i]
		}
	}
	return nil
}

type tile struct {
	x, y  int
	image string
}

func (t *tile) contains(x, y int) bool {
	return x >= t.x*tileSize && x < (t.x+1)*tileSize &&
		y >= t.y*tileSize && y < (t.y+1)*tileSize
}
