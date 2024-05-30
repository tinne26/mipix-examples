package main

import "image/color"

import "github.com/tinne26/mipix"
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/inpututil"

const GameWidth, GameHeight = 64, 48
const SquareSide = 4
const BackSquareSide = 2
const MoveSpeed = 0.06

type Game struct {
	backSquare *ebiten.Image
	square *ebiten.Image
	squareX float64
	squareY float64
}

func (self *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
		mipix.Redraw().Request()
	}

	// TPS change test
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		switch ebiten.TPS() {
		case 30: ebiten.SetTPS(240); mipix.Tick().SetRate(1)
		case 60: ebiten.SetTPS(30); mipix.Tick().SetRate(8)
		case 120: ebiten.SetTPS(60); mipix.Tick().SetRate(4)
		case 240: ebiten.SetTPS(120); mipix.Tick().SetRate(2)
		}
		mipix.Redraw().Request()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		switch ebiten.TPS() {
		case 30: ebiten.SetTPS(60); mipix.Tick().SetRate(4)
		case 60: ebiten.SetTPS(120); mipix.Tick().SetRate(2)
		case 120: ebiten.SetTPS(240); mipix.Tick().SetRate(1)
		case 240: ebiten.SetTPS(30); mipix.Tick().SetRate(8)
		}
		mipix.Redraw().Request()
	}

	// stretching mode
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		mipix.Scaling().SetStretchingAllowed(!mipix.Scaling().GetStretchingAllowed())
	}

	// trigger zoom
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		_, targetZoom := mipix.Camera().GetZoom()
		switch targetZoom {
		case 1.0: mipix.Camera().Zoom(3.0)
		case 3.0: mipix.Camera().Zoom(1.0)
		}
	}

	// trigger shake
	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		mipix.Camera().TriggerShake(mipix.ZeroTicks, 480, 160)
	}

	// update square position
	simulationRate := float64(mipix.Tick().GetRate())
	var xChange, yChange float64
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		xChange = -MoveSpeed*simulationRate
		mipix.Redraw().Request()
	case ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		xChange = +MoveSpeed*simulationRate
		mipix.Redraw().Request()
	}
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		yChange = -MoveSpeed*simulationRate
		mipix.Redraw().Request()
	case ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		yChange = +MoveSpeed*simulationRate
		mipix.Redraw().Request()
	}
	if xChange != 0 && yChange != 0 {
		xChange, yChange = xChange/1.4142, yChange/1.4142
	}
	self.squareX += xChange
	self.squareY += yChange
	mipix.Camera().NotifyCoordinates(self.squareX + SquareSide/2.0, self.squareY + SquareSide/2.0)
	return nil
}

func (self *Game) Draw(canvas *ebiten.Image) {
	if !mipix.Redraw().Pending() { return }
	rect := mipix.Camera().Area()

	// draw checkerboard pattern
	mod := func(a, b int) int { return (a % b + b) % b }
	var opts ebiten.DrawImageOptions
	for y := rect.Min.Y - mod(rect.Min.Y, BackSquareSide); y < rect.Max.Y; y += BackSquareSide {
		for x := rect.Min.X - mod(rect.Min.X, BackSquareSide); x < rect.Max.X; x += BackSquareSide {
			opts.GeoM.Translate(float64(x - rect.Min.X), float64(y - rect.Min.Y))
			if y/BackSquareSide & 0b1 != x/BackSquareSide & 0b1 {
				opts.ColorScale.Scale(0.9, 0.9, 0.9, 1.0)
			} else {
				opts.ColorScale.Scale(0.96, 0.96, 0.96, 1.0)
			}
			canvas.DrawImage(self.backSquare, &opts)
			opts.ColorScale.Reset()
			opts.GeoM.Reset()
		}
	}

	// main square draw
	mipix.QueueHiResDraw(self.DrawSquare)

	// info / debug draws
	mipix.Debug().Drawf("[R/T] Sim. rate (%dUPS/%dTPU)", mipix.Tick().UPS(), mipix.Tick().GetRate())
	mipix.Debug().Drawf("[F] Fullscreen")
	zoom, _ := mipix.Camera().GetZoom()
	mipix.Debug().Drawf("[Z] Zoom (x%.02f)", zoom)
	mipix.Debug().Drawf("[X] Shake")
	if mipix.Scaling().GetStretchingAllowed() {
		mipix.Debug().Drawf("[K] Stretch [ON]")
	} else {
		mipix.Debug().Drawf("[K] Stretch [OFF]")
	}
}

func (self *Game) DrawSquare(_, target *ebiten.Image) {
	mipix.HiRes().Draw(target, self.square, self.squareX, self.squareY)
}

func main() {
	ebiten.SetTPS(60)
	mipix.Tick().SetRate(4)
	ebiten.SetWindowTitle("mipix-examples/src/simrate")
	mipix.SetResolution(GameWidth, GameHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	mipix.Redraw().SetManaged(true)
	ebiten.SetScreenClearedEveryFrame(false)

	game := &Game{
		backSquare: ebiten.NewImage(BackSquareSide, BackSquareSide),
		square: ebiten.NewImage(SquareSide, SquareSide),
	}
	game.backSquare.Fill(color.RGBA{255, 255, 255, 255})
	game.square.Fill(color.RGBA{64, 255, 192, 255})

	// set camera initial position
	mipix.Camera().ResetCoordinates(SquareSide/2.0, SquareSide/2.0)

	// run the game
	err := mipix.Run(game)
	if err != nil { panic(err) }
}
