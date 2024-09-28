package main

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/tinne26/mipix"
import "github.com/tinne26/mipix/utils"

type Game struct {
	LookAtX, LookAtY float64
}

func (game *Game) Update() error {
	// detect directions
	up    := ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	down  := ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	left  := ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	right := ebiten.IsKeyPressed(ebiten.KeyArrowRight)
	if up   && down  { up  , down  = false, false }
	if left && right { left, right = false, false }
	
	// apply diagonal speed reduction if needed
	var speed float64 = 0.2
	if (up || down) && (left || right) {
		speed *= 0.7
	}

	// apply speed to camera target
	if up    { game.LookAtY -= speed }
	if down  { game.LookAtY += speed }
	if left  { game.LookAtX -= speed }
	if right { game.LookAtX += speed }

	// notify new camera target
	mipix.Camera().NotifyCoordinates(game.LookAtX, game.LookAtY)

	return nil
}

func (game *Game) Draw(canvas *ebiten.Image) {
	canvas.Fill(utils.RGB(255, 255, 255))
	camArea := mipix.Camera().Area()
	centerRect := utils.Rect(-1, -1, 1, 1)
	if centerRect.Overlaps(camArea) {
		drawRect := centerRect.Sub(camArea.Min)
		utils.FillOverRect(canvas, drawRect, utils.RGB(200, 0, 200))
	}
}

func main() {
	mipix.SetResolution(128, 72)
	err := mipix.Run(&Game{})
	if err != nil { panic(err) }
}
