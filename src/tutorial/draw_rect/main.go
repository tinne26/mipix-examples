package main

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/tinne26/mipix"
import "github.com/tinne26/mipix/utils"

const RectCX, RectCY = 0, 0 // place any desired coords here
var SomeRect = utils.Rect(RectCX - 1, RectCY - 1, RectCX + 1, RectCY + 1)
var RectColor = utils.RGBA(128, 128, 128, 128)

type Game struct {}

func (game *Game) Update() error {
	return nil
}

func (game *Game) Draw(canvas *ebiten.Image) {
	// white background fill
	canvas.Fill(utils.RGB(255, 255, 255))

	// fill the rect, which is defined in logical global coords
	camArea := mipix.Camera().Area()
	if SomeRect.Overlaps(camArea) {
		localRect := SomeRect.Sub(camArea.Min)
		utils.FillOverRect(canvas, localRect, RectColor)
	}
}

func main() {
	mipix.SetResolution(128, 72)
	err := mipix.Run(&Game{})
	if err != nil { panic(err) }
}
