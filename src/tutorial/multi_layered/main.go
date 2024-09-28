package main

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/tinne26/mipix"
import "github.com/tinne26/mipix/utils"

import "github.com/hajimehoshi/ebiten/v2/text/v2"
import "golang.org/x/image/font/opentype"
import "golang.org/x/image/font"
import "github.com/tinne26/fonts/liberation/lbrtsans"

// Helper type for decorative tiles
type Grass struct { X, Y int }
func (g Grass) Draw(canvas *ebiten.Image, cameraArea utils.Rectangle) {
	rect := utils.Rect(g.X*5, g.Y*5, g.X*5 + 5, g.Y*5 + 5)
	if rect.Overlaps(cameraArea) {
		fillRect := rect.Sub(cameraArea.Min)
		utils.FillOverRect(canvas, fillRect, utils.RGB(83, 141, 106))
	}
}

// Main game struct
type Game struct {
	PlayerCX, PlayerCY float64
	GrassTiles []Grass
	FontFace text.Face
	FontSize float64
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
	if up    { game.PlayerCY -= speed }
	if down  { game.PlayerCY += speed }
	if left  { game.PlayerCX -= speed }
	if right { game.PlayerCX += speed }

	// notify new camera target
	mipix.Camera().NotifyCoordinates(game.PlayerCX, game.PlayerCY)

	return nil
}

func (game *Game) Draw(canvas *ebiten.Image) {
	// fill background
	canvas.Fill(utils.RGB(128, 207, 169))

	// draw grass on the logical canvas
	cameraArea := mipix.Camera().Area()
	for _, grass := range game.GrassTiles {
		grass.Draw(canvas, cameraArea)
	}

	// queue draw for player rect at high resolution
	mipix.QueueHiResDraw(func(_, hiResCanvas *ebiten.Image) {
		ox, oy := game.PlayerCX - 1.5, game.PlayerCY - 1.5
		fx, fy := game.PlayerCX + 1.5, game.PlayerCY + 1.5
		rgba := utils.RGB(66, 67, 66)
		mipix.HiRes().FillOverRect(hiResCanvas, ox, oy, fx, fy, rgba)
	})

	// queue text rendering on high resolution too
	mipix.QueueHiResDraw(game.DrawText)
}

// High resolution text rendering function
func (game *Game) DrawText(_, hiResCanvas *ebiten.Image) {
	// determine text size
	bounds := hiResCanvas.Bounds()
	height := float64(bounds.Dy())
	fontSize := height/10.0

	// (re)initialize font face if necessary
	if game.FontSize != fontSize {
		var opts opentype.FaceOptions
		opts.DPI = 72.0
		opts.Size = fontSize
		opts.Hinting = font.HintingFull
		face, err := opentype.NewFace(lbrtsans.Font(), &opts)
		game.FontFace = text.NewGoXFace(face)
		game.FontSize = fontSize
		if err != nil { panic(err) }
	}

	// draw text
	var textOpts text.DrawOptions
	textOpts.PrimaryAlign = text.AlignCenter
	ox, oy := float64(bounds.Min.X), float64(bounds.Min.Y)
	textOpts.GeoM.Translate(ox + float64(bounds.Dx())/2.0, oy + (height - height/6.0))
	textOpts.ColorScale.ScaleWithColor(utils.RGB(30, 51, 39))
	textOpts.Blend = ebiten.BlendLighter
	text.Draw(hiResCanvas, "NOTHINGNESS AWAITS", game.FontFace, &textOpts)
}

func main() {
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	mipix.SetResolution(128, 72)
	err := mipix.Run(&Game{
		GrassTiles: []Grass{ // add some little decoration
			{-6, -6}, {8, -6}, {9, -6}, {-2, -5}, {4, -5}, {8, -5}, {-5, -4}, {-1, -4}, {2, -4},
			{-5, -3}, {-4, -3}, {-2, -3}, {-1, -3}, {1, -3}, {2, -3}, {3, -3}, 
			{-6, -2}, {-5, -2}, {-4, -2}, {-3, -2}, {-1, -2}, {0, -2}, {1, -2}, {2, -2}, {3, -2}, 
			{-4, -1}, {-3, -1}, {-2, -1}, {-1, -1}, {0, -1}, {1, -1}, {4, -1}, {-9, -1},
			{-5, 0}, {-3, 0}, {-2, 0}, {-1, 0}, {0, 0}, {1, 0}, {2, 0}, {-10, 0}, {-9, 0}, {-8, 0}, 
			{-3, 1}, {-2, 1}, {-1, 1}, {1, 1}, {2, 1}, {3, 1}, {-9, 1}, {-8, 1},
			{-2, 2}, {0, 2}, {1, 2}, {5, 2}, {-10, 2}, {-8, 2}, {-7, 2},
			{-3, 3}, {1, 3}, {3, 3}, {4, 3}, {5, 3}, {-8, 3},
			{3, 4}, {4, 4}, {-6, 5}, {11, 0}, {12, 1}, {12, 2},
		},
	})
	if err != nil { panic(err) }
}
