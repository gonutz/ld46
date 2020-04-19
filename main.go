package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gonutz/blob"
	"github.com/gonutz/payload"
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
		"assets/open_hand_cursor.png",
		"assets/closing_hand_cursor_1.png",
		"assets/closing_hand_cursor_2.png",
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
		tileSolid: "tile_solid",
		tileLeft:  "tile_left",
		tileRight: "tile_right",
		tileJump:  "tile_jump",
		tileDoor:  "tile_door",
		tileSpike: "tile_spike",
	}

	leftCue  = cue{speedX: -5}
	rightCue = cue{speedX: 5}
	jumpCue  = cue{speedY: -11}

	tilePreview = "tile_preview.png"

	tileMapping = map[rune]tile{
		'x': tile{kind: tileSolid, isSolid: true},
		'<': tile{kind: tileLeft, isSolid: true},
		'>': tile{kind: tileRight, isSolid: true},
		'^': tile{kind: tileJump, isSolid: true},
		'Z': tile{kind: tileJump, isSolid: true, isDraggable: true},
		'(': tile{kind: tileLeft, isSolid: true, isDraggable: true},
		')': tile{kind: tileRight, isSolid: true, isDraggable: true},
		'o': tile{kind: tileSolid, isSolid: true, isDraggable: true},
		'D': tile{kind: tileDoor},
		'|': tile{kind: tileSpike},
	}

	levels = []string{
		`
	.             Dx.
	.s             x.
	.>xxxxxxxxxxxxxx.
	`,

		`
	.             Dx.
	.s           o x.
	.>xxxxxxxxxxxxxx.
	`,

		`
	.             D .
	.s           o  .
	.>xxxxxxxxxxxxxx.
	`,

		`
	.s         ooo   .
	.                .
	.              Dx.
	.               x.
	.>xxxxxxxxx   xxx.
	.         x|||x  .
	.         xxxxx  .
	`,

		`
	.s         oo    .
	.                .
	.              Dx.
	.               x.
	.>xxxxxxxxx   xxx.
	.         x|||x  .
	.         xxxxx  .
	`,

		`
	.s          o    .
	.                .
	.              Dx.
	.               x.
	.>xxxxxxxxx   xxx.
	.         x|||x  .
	.         xxxxx  .
	`,

		`
	.s              x.
	.               x.
	.               x.
	.               x.
	.          Z    x.
	.              Dx.
	.               x.
	.>xxxxxxx     xxx.
	.       x|||||x  .
	.       xxxxxxx  .
	`,

		`
	.s              x.
	.               x.
	.               x.
	.               x.
	.          Z    x.
	.              Dx.
	.               x.
	.>xxxxxx      xxx.
	.      x||||||x  .
	.      xxxxxxxx  .
	`,

		`
	.s                .
	.                 .
	.                 .
	.                 .
	.          Z      .
	.              Dx .
	.               x .
	.>xxxxxx      xxx .
	.      x||||||x   .
	.      xxxxxxxx   .
	`,

		`
	.                   .
	.                   .
	.s                  .
	.>xx                .
	.x                  .
	.x                  .
	.x                  .
	.xD (               .
	.x                  .
	.xxxxxxxxxxxxxxxxx  .
	.                   .
	.                   .
	`,

		`
	.                   .
	.                   .
	.                   .
	.                 D .
	. s                 .
	. )xxxxxxxxx(xxxxxx .
	.                   .
	.                   .
	`,

		`
	.                     .
	.                     .
	.           (   )     .
	.                     .
	.                     .
	. s                   .
	. >xxxxxxxxxxx        .
	.                     .
	.                     .
	.              xxxxxx .
	.                     .
	.                     .
	.       xxxxxx        .
	.                     .
	.                     .
	.              xxxxxx .
	.       D             .
	.                     .
	.       x^^^^^        .
	.                     .
	.                     .
	`,

		`
	.           Z            .
	.                        .
	.                     Dx .
	.s                     x .
	.>xxxxxxxxxxo<<<<<<<xxxx .
	.          xxx           .
	`,

		`
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	.                                                  D .
	s            o                                       .
	>xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
	`,

		`
	.               .
	.               .
	. s             .
	. >             .
	. .             .
	. .             .
	. .   ^x^xx<    .
	. .             .
	. .             .
	. >    oo     D .
	. .             .
	. .           x .
	. .   x^x     x .
	. x|||||||||||x .
	. xxxxxxxxxxxxx .
	.               .
	.               .
	`,
	}
	levelIndex = 0
	levelLost  = 0 // Call isLevelLost() to see if the level was lost.
	lostTimer  = 0

	speedX    = 0
	speedY    = 0.0
	gravity   = 0.36
	maxSpeedY = 15.0
	player    = rect(0, 0, 48, 96)
	falling   = false
)

func isLevelLost() bool {
	return levelLost > 3
}

type dummyCloser struct {
	io.ReadSeeker
}

func (dummyCloser) Close() error { return nil }

func main() {
	// For the release we build all assets into the executable. This reads that
	// data and if it finds it, directs the prototype lib to use that instead of
	// the file system.
	if assets, err := payload.Open(); err == nil {
		defer assets.Close()
		if r, err := blob.Open(assets); err == nil {
			draw.OpenFile = func(path string) (io.ReadCloser, error) {
				f, found := r.GetByID(strings.TrimPrefix(path, "assets/"))
				if found {
					return dummyCloser{f}, nil
				} else {
					panic(path + " not found in blob")
					return nil, errors.New(path + " not found in blob")
				}
			}
		}
	}

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

	toggleFullscreen := func() {
		windowFullscreen = !windowFullscreen
		centerCamera = true
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
		if window.WasKeyPressed(draw.KeyF11) {
			toggleFullscreen()
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

		// Draw texts at the top: which level are we in, what keyboard controls
		// are there.
		const textScale = 1.6
		textY := 15
		textBackground := blue3
		textBackground.A = 0.5

		wasLeftClicked := false
		for _, c := range window.Clicks() {
			if c.Button == draw.LeftButton {
				wasLeftClicked = true
			}
		}

		{
			text := fmt.Sprintf("Level %d/%d", levelIndex+1, len(levels))
			textW, textH := window.GetScaledTextSize(text, textScale)
			textX := 5
			window.FillRect(textX-5, textY-5, textW+10, textH+10, textBackground)
			window.DrawScaledText(text, textX, textY, textScale, blue8)
		}

		{
			if isLevelLost() {
				lostTimer++
			}
			textScale := textScale + float32(30-abs(lostTimer%60-30))/15
			text := "Restart (F2)"
			textW, textH := window.GetScaledTextSize(text, textScale)
			textX := (windowW - textW) / 2
			window.FillRect(textX-5, textY-5, textW+10, textH+10, textBackground)
			if rect(
				textX-5, textY-5, textW+10, textH+10,
			).contains(screenX(mx), screenY(my)) {
				window.FillRect(textX-5, textY-5, textW+10, textH+10, blue2)
				window.DrawRect(textX-5, textY-5, textW+10, textH+10, blue1)
				if wasLeftClicked {
					startLevel()
				}
			}
			window.DrawScaledText(text, textX, textY, textScale, blue8)
		}

		{
			text := "Fullscreen (F11)"
			textW, textH := window.GetScaledTextSize(text, textScale)
			textX := windowW - textW - 5
			window.FillRect(textX-5, textY-5, textW+10, textH+10, textBackground)
			if rect(
				textX-5, textY-5, textW+10, textH+10,
			).contains(screenX(mx), screenY(my)) {
				window.FillRect(textX-5, textY-5, textW+10, textH+10, blue2)
				window.DrawRect(textX-5, textY-5, textW+10, textH+10, blue1)
				if wasLeftClicked {
					toggleFullscreen()
				}
			}
			window.DrawScaledText(text, textX, textY, textScale, blue8)
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
			if t, ok := tileMapping[r]; ok {
				t.x = x
				t.y = y
				l.tiles = append(l.tiles, t)
			}
			switch r {
			case 's':
				l.playerX, l.playerY = x, y
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
	// +9999 is to make sure modulo works for negative x and y values.
	if (x+9999*tileSize)%tileSize == tileSize/2 &&
		(y+9999*tileSize)%tileSize == 0 {
		t := l.tileAt(toTile(x), toTile(y))
		if t == nil {
			return nil
		}
		if t.kind == tileLeft {
			return &leftCue
		}
		if t.kind == tileRight {
			return &rightCue
		}
		if t.kind == tileJump {
			return &jumpCue
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
	s := "assets/" + tileTypeToImage[t.kind]
	if t.isDraggable {
		s += "_draggable"
	}
	if t.highlighted {
		s += "_highlight"
	}
	return s + ".png"
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

func (r rectangle) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
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
