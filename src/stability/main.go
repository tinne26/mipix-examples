package main

import "math"
import "embed"
import "image/png"
import "image/color"

import "github.com/tinne26/mipix"
import "github.com/tinne26/mipix/tracker"
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/inpututil"

//go:embed sword.png
var assets embed.FS

const GameWidth, GameHeight = 256, 144

// --- graphic ---

type Graphic struct {
	X, Y int
	Source *ebiten.Image
}

func (self *Graphic) Center() (float64, float64) {
	bounds := self.Source.Bounds()
	return float64(self.X) + float64(bounds.Dx())/2.0, float64(self.X) + float64(bounds.Dy())/2.0
}

func loadGraphic(name string, x, y int) Graphic {
	file, err := assets.Open(name + ".png")
	if err != nil { panic(err) }
	img, err := png.Decode(file)
	if err != nil { panic(err) }
	source := ebiten.NewImageFromImage(img)
	err = file.Close()
	if err != nil { panic(err) }
	return Graphic{ X: x, Y: y, Source: source }
}

// --- game ---

type Game struct {
	graphic Graphic
	swing bool
	xOffset float64
	needsRedraw bool
}

func (self *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	
	// scaling filter changes
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		mipix.Scaling().SetFilter((mipix.Scaling().GetFilter() + 1) % 9)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		mipix.Scaling().SetFilter((mipix.Scaling().GetFilter() + 8) % 9)
	}

	// swing and zoom changes
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		_, target := mipix.Camera().GetZoom()
		switch target {
		case 1.0: mipix.Camera().Zoom(2.0)
		case 2.0: mipix.Camera().Zoom(1.0)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		self.swing = !self.swing
		if !self.swing { self.xOffset = 0.0 }
	}

	// update swing and zoom
	if self.swing {
		self.xOffset = math.Sin(float64(mipix.Tick().Now())/40.0)*2.0
	}
	x, y := self.graphic.Center()
	mipix.Camera().NotifyCoordinates(x + self.xOffset, y)
	self.needsRedraw = true
	return nil
}

func (self *Game) Draw(canvas *ebiten.Image) {
	if !self.needsRedraw && !mipix.LayoutHasChanged() { return }
	self.needsRedraw = false

	mipix.Debug().Drawf("[Q/E] %s filter", mipix.Scaling().GetFilter().String())
	mipix.Debug().Drawf("[S] Swing on/off")
	mipix.Debug().Drawf("[Z] Zoom")
	mipix.Debug().Drawf("[F] Fullscreen")

	canvas.Fill(color.RGBA{244, 232, 232, 255})
	origin := mipix.Camera().Area().Min
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(float64(self.graphic.X - origin.X), float64(self.graphic.Y - origin.Y))
	canvas.DrawImage(self.graphic.Source, &opts)
}

func main() {
	ebiten.SetWindowTitle("mipix-examples/src/stability")
	mipix.SetResolution(GameWidth, GameHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(false)

	// create game and run it
	game := &Game{ graphic: loadGraphic("sword", 0, 0) }
	mipix.Camera().SetTracker(tracker.Linear)
	x, y := game.graphic.Center()
	mipix.Camera().ResetCoordinates(x, y)
	err := mipix.Run(game)
	if err != nil { panic(err) }
}
