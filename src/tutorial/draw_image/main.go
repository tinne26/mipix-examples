package main

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/tinne26/mipix"
import "github.com/tinne26/mipix/utils"

const PrawnOX, PrawnOY = -3, -3 // place any desired coords here
var Prawn *ebiten.Image = utils.MaskToImage(6, []uint8{
	0, 0, 0, 0, 1, 0, // example low-res image
	0, 0, 0, 0, 1, 1,
	0, 0, 1, 1, 0, 0,
	0, 1, 1, 1, 0, 0,
	1, 1, 1, 0, 0, 0,
	1, 1, 0, 0, 0, 0,
}, utils.RGB(219, 86, 32))

type Game struct {}

func (game *Game) Update() error {
	return nil
}

func (game *Game) Draw(canvas *ebiten.Image) {
	// set some background color
	canvas.Fill(utils.RGB(255, 255, 255))

	// obtain the camera area that mipix is requesting
	// us to draw. this is the most critical function
	// that has to be used when drawing with mipix
	camArea := mipix.Camera().Area()

	// see if our content overlaps the area we need to
	// draw, and if it does, we subtract the camera
	// origin coordinates to our object's global coords
	prawnGlobalRect := utils.Shift(Prawn.Bounds(), PrawnOX, PrawnOY)
	if prawnGlobalRect.Overlaps(camArea) {
		// translate from global to local (canvas) coordinates
		prawnLocalRect := prawnGlobalRect.Sub(camArea.Min)

		// create DrawImageOptions and apply draw position
		var opts ebiten.DrawImageOptions
		tx := prawnLocalRect.Min.X
		ty := prawnLocalRect.Min.Y
		opts.GeoM.Translate(float64(tx), float64(ty))
		canvas.DrawImage(Prawn, &opts)
	}
}

func main() {
	mipix.SetResolution(128, 72)
	err := mipix.Run(&Game{})
	if err != nil { panic(err) }
}
