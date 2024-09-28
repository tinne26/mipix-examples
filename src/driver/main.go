package main

import ("image/color" ; "math" ; "math/rand/v2")
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/inpututil"
import "github.com/tinne26/mipix"
import "github.com/tinne26/mipix/shaker"
import "github.com/tinne26/mipix/utils"

// A driving game example showcasing basic structure,
// camera tracking, zooms and shakes.

const GameWidth, GameHeight, RoadWidth = 128, 72, 16
var BackRGB, RoadRGB = utils.RGB(126, 224, 129), utils.RGB(25, 21, 22)
var WheelBarRGB, WheelPinRGB = utils.RGB(250, 246, 246), utils.RGB(8, 103, 136)
var VehicleRGB = utils.RGB(255, 22, 84)

// --- bézier curves for the track, you can ignore this ---

type Curve struct {
	ox, oy, fx, fy float64 // start and end points
	ocy, fcy float64 // bézier control point y's
}

func (self *Curve) Draw(canvas *ebiten.Image, clr color.Color) {
	area := mipix.Camera().Area()
	self.eachYLine(func(x float64, baseY int) {
		if baseY + 1 < area.Min.Y || baseY > area.Max.Y { return }
		xl := int(math.Round(x - RoadWidth/2.0)) - area.Min.X
		xr, yt := xl + RoadWidth, baseY - area.Min.Y
		utils.FillOverRect(canvas, utils.Rect(xl, yt, xr, yt + 1), clr)
	})
}

func (self *Curve) Reroll(ox, oy float64) {
	self.ox, self.oy = ox, oy
	self.fx = ox + 24.0*(rand.Float64() - 0.5)*2.0
	self.fy = oy - (GameHeight + 2 + math.Floor((GameHeight/3.0)*rand.Float64()))
	dy := self.oy - self.fy
	self.ocy = self.oy - (dy*0.3 + dy*rand.Float64()*0.45)
	self.fcy = self.fy + (dy*0.3 + dy*rand.Float64()*0.45)
}

func (self *Curve) GetClosestX(refY float64) float64 {
	const TimeStep = 0.003
	var bestX, minDist float64 = 0, 6666.0
	for t := 0.0; t < 0.99999 + TimeStep; t += TimeStep {
		x, y := self.eval(t) // just brute forcing
		dist := math.Abs(y - refY)
		if dist < minDist {
			bestX, minDist = x, dist
		}
	}
	return bestX
}

func (self *Curve) ContainsY(y float64) bool {
	return y >= self.fy && y <= self.oy
}

// evaluates the curve brute forcing it and yields the x
// values of the closest y's to .5 for each pixel line
func (self *Curve) eachYLine(yield func(float64, int)) {
	const TimeStep = 0.002
	aBaseY, aX, aBestDist := self.oy + 0.0, self.ox, 0.5
	bBaseY, bX, bBestDist := self.oy - 1.0, self.ox, 1.5
	for t := TimeStep; t < 0.99999 + TimeStep; t += TimeStep {
		x, y := self.eval(t)
		aDist := math.Abs((aBaseY - 0.5) - y)
		bDist := math.Abs((bBaseY - 0.5) - y)
		if aDist <= aBestDist {
			aX, aBestDist = x, aDist
			bX, bBestDist = x, bDist
		} else {
			yield(aX, int(aBaseY))
			aBaseY, bBaseY = bBaseY, bBaseY - 1.0
			aX, aBestDist = bX, bBestDist
			if bDist <= bBestDist {
				bX, bBestDist = x, bDist
			} else {
				panic("timestep too coarse")
			}
		}
	}
	yield(aX, int(aBaseY))
}

func (self *Curve) eval(t float64) (float64, float64) {
	oc1x , oc1y  := self.ox, lerp(self.oy, self.ocy, t) // origin to control 1
	c2fx , c2fy  := self.fx, lerp(self.fcy, self.fy, t)  // control 2 to end
	c1c2x, c1c2y := lerp2(self.ox, self.ocy, self.fx, self.fcy, t) // control 1 to control 2
	iox  , ioy   := lerp2(oc1x, oc1y, c1c2x, c1c2y, t) // first interpolation from origin
	ifx  , ify   := lerp2(c1c2x, c1c2y, c2fx, c2fy, t) // second interpolation to end
	return lerp2(iox, ioy, ifx, ify, t) // cubic interpolation
}
func lerp2(ax, ay, bx, by float64, t float64) (float64, float64) {
	return lerp(ax, bx, t), lerp(ay, by, t)
}
func lerp(a, b float64, t float64) float64 {
	return a + t*(b - a)
}

// --- main game logic ---

type Game struct {
	ui *mipix.Offscreen
	vehicleCX, vehicleCY float64 // center X, centerY
	wheel float64
	track [2]Curve
}

func (self *Game) Update() error {
	// fullscreen mode
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	// turn left / right
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		self.wheel = max(self.wheel - 0.01, -1.0)
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		self.wheel = min(self.wheel + 0.01,  1.0)
	}

	// update vehicle position
	degrees := 90.0 - self.wheel*45.0
	dy, dx  := math.Sincos(degrees*math.Pi/180.0)
	self.vehicleCX += dx*0.4
	self.vehicleCY -= dy*0.23

	// notify new camera position
	mipix.Camera().NotifyCoordinates(self.vehicleCX, self.vehicleCY - GameHeight/6)

	// detect out of road and trigger shake
	t := 0 // active track index
	if self.track[1].ContainsY(self.vehicleCY) { t = 1 }
	xOffset := math.Abs(self.track[t].GetClosestX(self.vehicleCY) - self.vehicleCX)
	switch xOffset + 2.5 > RoadWidth/2.0 {
	case true  : mipix.Camera().StartShake(20) // going off road
	case false : mipix.Camera().EndShake(20)   // staying within road
	}

	// reroll tracks as needed
	currentZoom, targetZoom := mipix.Camera().GetZoom()
	area := mipix.Camera().Area()
	cutoffY := float64(area.Max.Y) + float64(area.Dy())*(currentZoom - 1.0)
	if self.track[0].fy > cutoffY {
		self.track[0].Reroll(self.track[1].fx, self.track[1].fy)
	} else if self.track[1].fy > cutoffY {
		self.track[1].Reroll(self.track[0].fx, self.track[0].fy)
	}

	// change zoom every once in a while
	if mipix.Tick().Now() % (60*10) == 0 {
		switch targetZoom {
		case 1.0 : mipix.Camera().Zoom(1.4)
		default  : mipix.Camera().Zoom(1.0)
		}
	}

	return nil
}

func (self *Game) Draw(canvas *ebiten.Image) {
	// draw background and track
	utils.FillOver(canvas, BackRGB)
	self.track[0].Draw(canvas, RoadRGB)
	self.track[1].Draw(canvas, RoadRGB)

	// draw car at high resolution coordinates
	mipix.QueueHiResDraw(self.DrawCarHiRes)
	
	// draw wheel/steering indicator
	self.ui.Clear()
	wcx, wcy := GameWidth/2, GameHeight - GameHeight/8
	wheelBarRect := utils.Rect(wcx - 8, wcy - 1, wcx + 8, wcy + 1)
	self.ui.CoatRect(wheelBarRect, WheelBarRGB)
	px := wcx + int(math.Round(self.wheel*8.0))
	wheelPinRect := utils.Rect(px - 1, wcy - 2, px + 1, wcy + 2)
	self.ui.CoatRect(wheelPinRect, WheelPinRGB)
	mipix.QueueHiResDraw(func(_, hiResCanvas *ebiten.Image) {
		self.ui.Project(hiResCanvas)
	})

	// print instructions and actions
	mipix.Debug().Drawf("[LEFT/RIGHT] Steer")
	mipix.Debug().Drawf("[F] Fullscreen")
}

func (self *Game) DrawCarHiRes(_, hiResCanvas *ebiten.Image) {
	xl, xr := self.vehicleCX - 3.0, self.vehicleCX + 3.0
	yt, yb := self.vehicleCY - 4.0, self.vehicleCY + 4.0
	mipix.HiRes().FillOverRect(hiResCanvas, xl, yt, xr, yb, VehicleRGB)
}

func main() {
	ebiten.SetWindowTitle("mipix-examples/src/driver")
	mipix.SetResolution(GameWidth, GameHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// configure off-road shaker
	var offRoadShaker shaker.Spring
	offRoadShaker.SetMotionScale(0.01, 0.007)
	offRoadShaker.SetParameters(0.1, 32.0)
	offRoadShaker.SetZoomCompensation(0.5)
	mipix.Camera().SetShaker(&offRoadShaker)

	// create and run the game
	game := Game{ ui: mipix.NewOffscreen(GameWidth, GameHeight) }
	game.track[0].Reroll(0, GameHeight/5.0)
	game.track[0].fx = 0
	game.track[1].Reroll(0, game.track[0].fy)
	err := mipix.Run(&game)
	if err != nil { panic(err) }
}
