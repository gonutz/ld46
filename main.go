package main

import (
	"fmt"
	"strings"

	"github.com/gonutz/prototype/draw"
)

var (
	windowTitle      = "LD46"
	windowFullscreen = true
	debugKeys        = true

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
	cameraSpeed      = 15
	// cameraMoveMargin is how close you have to be to the edge of the screen to
	// move the camera.
	cameraMoveMargin = 35
	centerCamera     = false

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

	tileSolid    tileType = 1
	tileLeft     tileType = 2
	tileRight    tileType = 3
	tileJump     tileType = 4
	tileDrag     tileType = 5
	tileJumpDrag tileType = 6
	tileDoor     tileType = 7
	tileSpike    tileType = 8

	tileTypeToImage = map[tileType]string{
		tileSolid:    "tile_solid.png",
		tileLeft:     "tile_left.png",
		tileRight:    "tile_right.png",
		tileJump:     "tile_jump.png",
		tileDrag:     "tile_draggable.png",
		tileJumpDrag: "tile_jump_draggable.png",
		tileDoor:     "tile_door.png",
		tileSpike:    "tile_spike.png",
	}

	tilePreview = "tile_preview.png"

	tileMapping = map[rune]tileType{
		'x': tileSolid,
		'<': tileLeft,
		'>': tileRight,
		'^': tileJump,
		'Z': tileJumpDrag,
		'o': tileDrag,
		'D': tileDoor,
		'|': tileSpike,
	}

	levels = []string{
		`
	xs              x
	x               x
	x               x
	x              Dx
	x               x
	x>xxxxxxxxxxxxxxx
	`,

		`
	xs              x
	x               x
	x               x
	x              Dx
	x       o       x
	x>xxxxxxxxxxxxxxx
	`,

		`
	xs              x
	x               x
	x               x
	x            D  x
	x       o       x
	x>xxxxxxxxxxxxxxx
	`,

		`
	x                     x
	x          o          x
	x                     x
	x                    Dx
	xs                    x
	x>xxxxxxxxx xxxxxxxxxxx
	.         x|x         .
	.         xxx         .
	`,

		`
	x                     x
	x          Z          x
	x                     x
	x                    Dx
	xs                    x
	x>xxxxxxxxxo<xxxxxxxxxx
	.         xxx         .
	`,

		`
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	.                                                  D.
	s            o                                     ..
	>xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	`,

		`
	s           .
	>           .
	.           .
	.           .
	.  <        .
	.           .
	.           .
	>   oo      .
	.           D
	.           .
	.  x^x      x
	`,
	}
	levelIndex = 0
	levelLost  = 0 // Call isLevelLost() to see if the level was lost.
	lostTimer  = 0

	speedX    = 0
	speedY    = 0.0
	gravity   = 0.35
	maxSpeedY = 10.0
	player    = rect(0, 0, 48, 96)
	falling   = false
)

func isLevelLost() bool {
	return levelLost > 3
}

func main() {
	var level *level
	startLevel := func() {
		level = parseLevel(levels[levelIndex])
		centerCamera = true
		player.x = level.playerX*tileSize + (tileSize-player.w)/2
		player.y = level.playerY*tileSize + tileSize - player.h
		handClosing = false
		leftMouseWasDown = false
		movingTile = nil
		levelLost = 0
		lostTimer = 0
		speedX = 0
		speedY = 0.0
		falling = false
	}

	previousLevel := func() {
		levelIndex = clamp(levelIndex-1, 0, len(levels)-1)
		level = nil
	}
	nextLevel := func() {
		levelIndex = clamp(levelIndex+1, 0, len(levels)-1)
		level = nil
	}

	err := draw.RunWindow(windowTitle, 1000, 600, func(window draw.Window) {
		// Some keys are handled right on top before the frame gets drawn, to be
		// most responsive.

		// Close window right away when requested.
		if window.WasKeyPressed(draw.KeyEscape) {
			window.Close()
			return
		}

		// Toggle fullscreen and center the camera on the player afterwards.
		if window.WasKeyPressed(draw.KeyEnter) &&
			(window.IsKeyDown(draw.KeyLeftAlt) || window.IsKeyDown(draw.KeyRightAlt)) {
			windowFullscreen = !windowFullscreen
			centerCamera = true
		}
		window.SetFullscreen(windowFullscreen)
		windowW, windowH := window.Size()
		if centerCamera {
			cameraX = player.x + player.w/2 - windowW/2
			cameraY = player.y + player.h/2 - windowH/2
			centerCamera = false
		}

		if window.WasKeyPressed(draw.KeyF2) {
			startLevel()
		}

		if debugKeys {
			if window.WasKeyPressed(draw.KeyLeft) {
				previousLevel()
			}
			if window.WasKeyPressed(draw.KeyRight) {
				nextLevel()
			}
		}

		// If we are at the door, start the next level.
		// TODO There should be an animation going into the door and then the
		// screen fades to black and comes back out of black in the next level.
		if level != nil && speedX == 0 && speedY == 0 {
			t := level.tileAt(toTile(player.x+player.w/2), toTile(player.y))
			if t != nil && t.kind == tileDoor {
				nextLevel()
			}
			levelLost++
		}

		if speedX != 0 || speedY != 0 {
			levelLost = 0
		}

		// Make sure we have a level right now.
		if level == nil {
			startLevel()
		}

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

		// Move the player.
		playerOnGround := func() bool {
			if speedY < 0 {
				// Jumping up means we are not on the ground, even if we still
				// have the floor under us because we are just starting to jump.
				return false
			}
			// NOTE Here we assume that the player is not wider than a tile.
			// If the player gets wider than a tile we have to compare more
			// tiles in x.
			tx1 := toTile(player.x)
			tx2 := toTile(player.x + player.w - 1)
			ty := toTile(player.y + player.h)
			return level.tileAt(tx1, ty).solid() || level.tileAt(tx2, ty).solid()
		}
		playerHitCeiling := func() bool {
			// NOTE Here we assume that the player is not wider than a tile.
			// If the player gets wider than a tile we have to compare more
			// tiles in x.
			tx1 := toTile(player.x)
			tx2 := toTile(player.x + player.w - 1)
			ty := toTile(player.y - 1)
			return level.tileAt(tx1, ty).solid() || level.tileAt(tx2, ty).solid()
		}
		yDist := round(speedY)
		doneInY := func() bool { return yDist == 0 }
		move1inY := func() {
			dy := sign(yDist)
			player.y += dy
			yDist -= dy
			if playerOnGround() {
				falling = false
				yDist = 0
				speedY = 0
			}
			if playerHitCeiling() {
				falling = true
				player.y++
				yDist = 0
				speedY = 0
			}
		}
		fall := func() {
			if !falling && !playerOnGround() {
				// Increment y so we hit a tile if we move in x direction
				// very fast.
				player.y++
				falling = true
			}
		}
		handleCues := func() {
			cue := level.cueAt(player.x+player.w/2, player.y+player.h)
			speedX, speedY = cue.updateSpeed(speedX, speedY)
		}
		func() {
			dx := sign(speedX)
			for i := 0; i < abs(speedX); i++ {
				move1inY()
				handleCues()

				player.x += dx
				for _, t := range level.tiles {
					if t.solid() && overlap(player, t.bounds()) {
						player.x -= dx
						speedX = 0
						return
					}
				}
			}
		}()
		for !doneInY() {
			move1inY()
		}
		fall()
		handleCues()

		if falling {
			speedY += gravity
			if speedY > maxSpeedY {
				speedY = maxSpeedY
			}
		}

		mx, my := world(window.MousePosition())
		leftMouseDown := window.IsMouseDown(draw.LeftButton)

		// If the player clicks on a draggable tile, start moving it.
		if !leftMouseWasDown && leftMouseDown {
			for i, t := range level.tiles {
				if t.draggable() && t.contains(mx, my) {
					movingTile = &level.tiles[i]
					movingTile.highlighted = true
					previewDx = mx - t.x*tileSize
					previewDy = my - t.y*tileSize
				}
			}
		}

		// If the player just stopped moving a tile, reset it.
		if !leftMouseDown && movingTile != nil {
			movingTile.highlighted = false
			movingTile = nil
		}

		// Move the currently dragged tile to the new mouse position.
		if movingTile != nil {
			x, y := mx-previewDx, my-previewDy
			newX := (x + sign(x)*tileSize/2) / tileSize
			newY := (y + sign(y)*tileSize/2) / tileSize
			t := level.tileAt(newX, newY)
			if (t == nil || t == movingTile) &&
				!overlap(player, rect(newX*tileSize, newY*tileSize, tileSize, tileSize)) {
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
			if movingTile == nil && t.draggable() && t.contains(mx, my) {
				t.highlighted = true
			}
			window.DrawImageFile(
				t.image(),
				screenX(t.x*tileSize),
				screenY(t.y*tileSize),
			)
		}

		// Draw player. TODO Have a real animation for the player.
		window.FillRect(
			screenX(player.x),
			screenY(player.y),
			player.w,
			player.h,
			blue5,
		)
		window.FillRect(
			screenX(player.x)+4,
			screenY(player.y)+4,
			player.w-8,
			player.h-8,
			blue1,
		)

		// Draw preview of the tile in movement.
		if movingTile != nil {
			window.DrawImageFile(
				tilePreview,
				screenX(mx)-previewDx,
				screenY(my)-previewDy,
			)
		}

		// Animate the hand opening/closing.
		if leftMouseDown {
			handFrame.inc()
		} else {
			handFrame.dec()
		}

		// Draw mouse cursor.
		window.DrawImageFile(
			handCursors[handFrame.value()],
			screenX(mx)-20,
			screenY(my)-20,
		)

		// Draw texts at the top: which level are we in, what keyboard controls
		// are there.
		const textScale = 1.6
		textBackground := blue3
		textBackground.A = 0.5

		{
			text := fmt.Sprintf("Level %d/%d", levelIndex+1, len(levels))
			textW, textH := window.GetScaledTextSize(text, textScale)
			textX := 5
			window.FillRect(textX-5, 0, textW+10, textH+10, textBackground)
			window.DrawScaledText(text, textX, 5, textScale, blue8)
		}

		{
			if isLevelLost() {
				lostTimer++
			}
			textScale := textScale + float32(30-abs(lostTimer%60-30))/15
			text := "F2: restart level"
			textW, textH := window.GetScaledTextSize(text, textScale)
			textX := (windowW - textW) / 2
			window.FillRect(textX-5, 0, textW+10, textH+10, textBackground)
			window.DrawScaledText(text, textX, 5, textScale, blue8)
		}

		{
			text := "Alt+Enter: toggle fullscreen"
			textW, textH := window.GetScaledTextSize(text, textScale)
			textX := windowW - textW - 5
			window.FillRect(textX-5, 0, textW+10, textH+10, textBackground)
			window.DrawScaledText(text, textX, 5, textScale, blue8)
		}

		// Update the frame information. These should always be at the end of
		// the frame.
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
	var l level
	y := 0
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if l.width == 0 {
			l.width = len(line)
		}
		if len(line) != l.width {
			panic("all lines in the level must have the same width")
		}
		for x, r := range line {
			x := x
			if kind, ok := tileMapping[r]; ok {
				l.tiles = append(l.tiles, tile{
					x:           x,
					y:           y,
					kind:        kind,
					isSolid:     !(kind == tileSpike || kind == tileDoor),
					isDraggable: kind == tileDrag || kind == tileJumpDrag,
				})
			}
			switch r {
			case 's':
				l.playerX, l.playerY = x, y
			case '<':
				l.cues = append(l.cues, cue{
					x:      x*tileSize + tileSize/2,
					y:      y * tileSize,
					speedX: -3,
					speedY: 0.0,
				})
			case '>':
				l.cues = append(l.cues, cue{
					x:      x*tileSize + tileSize/2,
					y:      y * tileSize,
					speedX: 3,
					speedY: 0.0,
				})
			case '^':
				l.cues = append(l.cues, cue{
					x:      x*tileSize + tileSize/2,
					y:      y * tileSize,
					speedX: 0,
					speedY: -14.0,
				})
			}
		}
		y++
	}
	l.width *= tileSize
	l.height = y * tileSize
	return &l
}

type level struct {
	tiles   []tile
	cues    []cue
	width   int
	height  int
	playerX int
	playerY int
}

func (l *level) tileAt(x, y int) *tile {
	for i, t := range l.tiles {
		if x == t.x && y == t.y {
			return &l.tiles[i]
		}
	}
	return nil
}

func (l *level) cueAt(x, y int) *cue {
	for i, c := range l.cues {
		if x == c.x && y == c.y {
			return &l.cues[i]
		}
	}
	return nil
}

type tile struct {
	x, y        int
	kind        tileType
	isSolid     bool
	isDraggable bool
	highlighted bool
}

type tileType int

func (t *tile) image() string {
	s := tileTypeToImage[t.kind]
	if t.highlighted {
		s = strings.TrimSuffix(s, ".png") + "_highlight.png"
	}
	return s
}

func (t *tile) contains(x, y int) bool {
	return x >= t.x*tileSize && x < (t.x+1)*tileSize &&
		y >= t.y*tileSize && y < (t.y+1)*tileSize
}

func (t *tile) bounds() rectangle {
	return rect(t.x*tileSize, t.y*tileSize, tileSize, tileSize)
}

func (t *tile) solid() bool {
	return t != nil && t.isSolid
}

func (t *tile) draggable() bool {
	return t != nil && t.isDraggable
}

type cue struct {
	x, y   int
	speedX int
	speedY float64
}

func (c *cue) updateSpeed(dx int, dy float64) (int, float64) {
	if c == nil {
		return dx, dy
	}
	if c.speedX != 0 {
		return c.speedX, 0
	}
	if c.speedY != 0 {
		return dx, c.speedY
	}
	return dx, dy
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
	if x > 0 {
		return 1
	}
	return 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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

type rectangle struct {
	x, y, w, h int
}

func rect(x, y, w, h int) rectangle {
	return rectangle{x: x, y: y, w: w, h: h}
}

func overlap(a, b rectangle) bool {
	return a.x < b.x+b.w && b.x < a.x+a.w &&
		a.y < b.y+b.h && b.y < a.y+a.h
}

func toTile(coord int) int {
	return coord / tileSize
}

func round(x float64) int {
	if x < 0 {
		return int(x - 0.5)
	}
	return int(x + 0.5)
}
