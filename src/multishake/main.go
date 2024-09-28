package main

import "image/color"

import "github.com/tinne26/mipix"
import "github.com/tinne26/mipix/shaker"
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/inpututil"

const GameWidth, GameHeight = 128, 72
const BackSquareSide = 2

// Constants for our shake channels.
const (
	ChanDefault shaker.Channel = iota // soft breathing/ship motion
	ChanSearch // anxiously looking for waldo or playing air hockey
	ChanTrigger // temporary aggressive random shaking
)

type Game struct {
	backSquare *ebiten.Image
}

func (self *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	// update zoom
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		_, targetZoom := mipix.Camera().GetZoom()
		if targetZoom == 1.0 {
			mipix.Camera().Zoom(2.0)
		} else {
			mipix.Camera().Zoom(1.0)
		}
	}

	// stop/start background shake
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		if mipix.Camera().IsShaking(ChanDefault) {
			mipix.Camera().EndShake(60, ChanDefault)
		} else {
			mipix.Camera().StartShake(60, ChanDefault)
		}
	}

	// stop/start search shaking
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if mipix.Camera().IsShaking(ChanSearch) {
			mipix.Camera().EndShake(60, ChanSearch)
		} else {
			mipix.Camera().StartShake(60, ChanSearch)
		}
	}

	// trigger temporary shake
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		mipix.Camera().TriggerShake(0, 90, 60, ChanTrigger)
	}

	return nil
}

func (self *Game) Draw(canvas *ebiten.Image) {
	// draw checkerboard pattern
	rect := mipix.Camera().Area()
	mod := func(a, b int) int { return (a % b + b) % b }
	var opts ebiten.DrawImageOptions
	for y := rect.Min.Y - mod(rect.Min.Y, BackSquareSide); y < rect.Max.Y; y += BackSquareSide {
		for x := rect.Min.X - mod(rect.Min.X, BackSquareSide); x < rect.Max.X; x += BackSquareSide {
			opts.GeoM.Translate(float64(x - rect.Min.X), float64(y - rect.Min.Y))
			if y/BackSquareSide & 0b1 != x/BackSquareSide & 0b1 {
				opts.ColorScale.Scale(0.86, 0.86, 0.86, 1.0)
			} else {
				opts.ColorScale.Scale(0.96, 0.96, 0.96, 1.0)
			}
			canvas.DrawImage(self.backSquare, &opts)
			opts.ColorScale.Reset()
			opts.GeoM.Reset()
		}
	}

	// instructions and actions
	mipix.Debug().Drawf("[F] Fullscreen")
	mipix.Debug().Drawf("[Z] Zoom")
	if mipix.Camera().IsShaking(ChanDefault) {
		mipix.Debug().Drawf("[B] Stop Back Shake")
	} else {
		mipix.Debug().Drawf("[B] Start Back Shake")
	}
	if mipix.Camera().IsShaking(ChanSearch) {
		mipix.Debug().Drawf("[D] Stop Search Shake")
	} else {
		mipix.Debug().Drawf("[D] Start Search Shake")
	}
	mipix.Debug().Drawf("[S] Trigger Shake")	
}

func main() {
	ebiten.SetWindowTitle("mipix-examples/src/multishake")
	mipix.SetResolution(GameWidth, GameHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// configure all shakers
	var backShaker shaker.Spring
	backShaker.SetMotionScale(0.014, 0.014)
	backShaker.SetParameters(0.15, 1.5)
	backShaker.SetZoomCompensation(1.0)
	mipix.Camera().SetShaker(&backShaker, ChanDefault)
	
	var searchShaker shaker.Quake
	searchShaker.SetMotionScale(0.15)
	searchShaker.SetSpeedRange(0.6, 1.8)
	searchShaker.SetZoomCompensation(0.5)
	mipix.Camera().SetShaker(&searchShaker, ChanSearch)

	var triggerShaker shaker.Random
	triggerShaker.SetMotionScale(0.026)
	mipix.Camera().SetShaker(&triggerShaker, ChanTrigger)

	// run the game
	game := Game{ backSquare: ebiten.NewImage(BackSquareSide, BackSquareSide) }
	game.backSquare.Fill(color.RGBA{220, 20, 60, 255})
	err := mipix.Run(&game)
	if err != nil { panic(err) }
}
