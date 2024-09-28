package main

import "image/color"

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/tinne26/mipix"

type Game struct {
	HiCursorX, HiCursorY int
	LoCursorX, LoCursorY float64
	LoRelativeX, LoRelativeY float64
	LoGameX, LoGameY float64
}

func (game *Game) Update() error {
	hiX, hiY := ebiten.CursorPosition()
	loX, loY := mipix.Convert().ToLogicalCoords(hiX, hiY)
	reX, reY := mipix.Convert().ToRelativeCoords(hiX, hiY)
	gmX, gmY := mipix.Convert().ToGameResolution(hiX, hiY)
	game.HiCursorX, game.HiCursorY = hiX, hiY
	game.LoCursorX, game.LoCursorY = loX, loY
	game.LoRelativeX, game.LoRelativeY = reX, reY
	game.LoGameX, game.LoGameY = gmX, gmY
	return nil
}

func (game *Game) Draw(canvas *ebiten.Image) {
	canvas.Fill(color.RGBA{128, 128, 128, 255})
	mipix.Debug().Drawf("[ Cursor Position ]")
	mipix.Debug().Drawf("High-res screen: (%d, %d)", game.HiCursorX, game.HiCursorY)
	mipix.Debug().Drawf("Low-res relative: (%.02f, %.02f)", game.LoRelativeX, game.LoRelativeY)
	mipix.Debug().Drawf("Low-res screen: (%.02f, %.02f)", game.LoGameX, game.LoGameY)
	mipix.Debug().Drawf("Low-res global: (%.02f, %.02f)", game.LoCursorX, game.LoCursorY)
}

func main() {
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	mipix.SetResolution(100, 100)
	err := mipix.Run(&Game{})
	if err != nil { panic(err) }
}
