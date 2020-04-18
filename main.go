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

	cameraX, cameraY = -50, -200
	cameraSpeed      = 10
	// cameraMoveMargin is how close you have to be to the edge of the screen to
	// move the camera.
	cameraMoveMargin = 35

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

	tileSolid     = "solid_tile.png"
	tileDrag      = "draggable_tile.png"
	tileDoor      = "door_tile.png"
	tileSpike     = "tile_spike.png"
	tileHighlight = "highlighted_tile.png"
	tilePreview   = "preview_tile.png"

	tileMapping = map[rune]string{
		'x': tileSolid,
		'o': tileDrag,
		'D': tileDoor,
		'^': tileSpike,
	}

	level1 = parseLevel(`
	.          ooo          .
	.                       .
	.                       D
	.                       .
	xxxxxxxxxxx   xxxxxxxxxxx
	.         x^^^x         .
	.         xxxxx         .
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

		// Clamp the camera to the level boundaries.
		cameraX = clamp(cameraX, 0, level.width-windowW)
		cameraY = clamp(cameraY, 0, level.height-windowH)

		// If the level is smaller than the screen, put it in the center.
		if windowW >= level.width {
			cameraX = (level.width - windowW) / 2
		}
		if windowH >= level.height {
			cameraY = (level.height - windowH) / 2
		}

		mx, my := world(window.MousePosition())
		leftMouseDown := window.IsMouseDown(draw.LeftButton)

		// Animate the hand opening/closing.
		if leftMouseDown {
			handFrame.inc()
		} else {
			handFrame.dec()
		}

		// If the player clicks on a draggable tile, start moving it.
		if !leftMouseWasDown && leftMouseDown {
			for i, t := range level.tiles {
				if t.image == tileDrag && t.contains(mx, my) {
					movingTile = &level.tiles[i]
					movingTile.image = tileHighlight
					previewDx = mx - t.x*tileSize
					previewDy = my - t.y*tileSize
				}
			}
		}

		// If the player just stopped moving a tile, reset it.
		if !leftMouseDown && movingTile != nil {
			movingTile.image = tileDrag
			movingTile = nil
		}

		// Move the currently dragged tile to the new mouse position.
		if movingTile != nil {
			x, y := mx-previewDx, my-previewDy
			newX := (x + sign(x)*tileSize/2) / tileSize
			newY := (y + sign(y)*tileSize/2) / tileSize
			t := level.tileAt(newX, newY)
			if t == nil || t == movingTile {
				movingTile.x = newX
				movingTile.y = newY
			}
		}

		var cameraDx, cameraDy int
		if screenX(mx) <= cameraMoveMargin {
			cameraDx = -cameraSpeed
		}
		if screenX(mx) >= windowW-1-cameraMoveMargin {
			cameraDx = cameraSpeed
		}
		if screenY(my) <= cameraMoveMargin {
			cameraDy = -cameraSpeed
		}
		if screenY(my) >= windowH-1-cameraMoveMargin {
			cameraDy = cameraSpeed
		}

		// Draw background sky as light blue gradient.
		for y := 0; y < windowH; y++ {
			window.DrawLine(
				0, y, windowW, y,
				lerpColor(blue1, blue3, float32(y)/float32(windowH-1)),
			)
		}

		// Draw tiles.
		for _, t := range level.tiles {
			image := t.image
			if movingTile == nil && t.image == tileDrag && t.contains(mx, my) {
				image = tileHighlight
			}
			window.DrawImageFile(image, screenX(t.x*tileSize), screenY(t.y*tileSize))
		}

		// Draw mouse cursor.
		window.DrawImageFile(handCursors[handFrame.value()], screenX(mx)-20, screenY(my)-20)
		if movingTile != nil {
			window.DrawImageFile(tilePreview, screenX(mx)-previewDx, screenY(my)-previewDy)
		}

		leftMouseWasDown = leftMouseDown
		cameraX += cameraDx
		cameraY += cameraDy
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
	var tiles []tile
	y := 0
	levelWidth := 0
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if levelWidth == 0 {
			levelWidth = len(line)
		}
		if len(line) != levelWidth {
			panic("all lines in the level must have the same width")
		}
		for x, r := range line {
			x := x
			if image, ok := tileMapping[r]; ok {
				tiles = append(tiles, tile{x: x, y: y, image: image})
			}
		}
		y++
	}
	return &level{
		tiles:  tiles,
		width:  levelWidth * tileSize,
		height: y * tileSize,
	}
}

type level struct {
	tiles  []tile
	width  int
	height int
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

func worldX(screenX int) int {
	return screenX + cameraX
}

func worldY(screenY int) int {
	return screenY + cameraY
}

func world(screenX, screenY int) (int, int) {
	return worldX(screenX), worldY(screenY)
}

func screenX(worldX int) int {
	return worldX - cameraX
}

func screenY(worldY int) int {
	return worldY - cameraY
}

func screen(worldX, worldY int) (int, int) {
	return screenX(worldX), screenY(worldY)
}

func sign(x int) int {
	if x < 0 {
		return -1
	}
	return 1
}

func clamp(x, min, max int) int {
	if x < min {
		x = min
	}
	if x > max {
		x = max
	}
	return x
}
