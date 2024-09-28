package main

import "embed"
import "image"
import "image/png"
import "image/color"

import "github.com/tinne26/mipix"
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/inpututil"

//go:embed assets/*
var assets embed.FS // see assets/README.md for licensing

const GameWidth, GameHeight = 256, 144

// --- graphic ---

type Graphic struct {
	X, Y int
	Source *ebiten.Image
}

var graphicSources map[string]*ebiten.Image = make(map[string]*ebiten.Image, 32)
func loadGraphic(name string, x, y int) Graphic {
	source, found := graphicSources[name]
	if !found {
		file, err := assets.Open("assets/" + name + ".png")
		if err != nil { panic(err) }
		img, err := png.Decode(file)
		if err != nil { panic(err) }
		source = ebiten.NewImageFromImage(img)
		err = file.Close()
		if err != nil { panic(err) }
		graphicSources[name] = source
	}
	return Graphic{ X: x, Y: y, Source: source }
}

// --- animation ---

type Animation struct {
	frames []*ebiten.Image
	frameDurations []uint8
	frameDurationLeft uint8
	frameIndex, loopIndex uint8
}
func (self *Animation) AddFrame(frame *ebiten.Image, durationTicks uint8) {
	if durationTicks == 0 { panic("durationTicks == 0") }
	self.frames = append(self.frames, frame)
	self.frameDurations = append(self.frameDurations, durationTicks)
	if len(self.frames) == 1 { self.frameDurationLeft = durationTicks }
}
func (self *Animation) Update() {
	self.frameDurationLeft -= 1
	if self.frameDurationLeft == 0 {
		if self.frameIndex == uint8(len(self.frames) - 1) {
			self.frameIndex = self.loopIndex
		} else {
			self.frameIndex += 1
		}
		self.frameDurationLeft = self.frameDurations[self.frameIndex]
	}
}
func (self *Animation) GetFrame() *ebiten.Image {
	return self.frames[self.frameIndex]
}
func (self *Animation) InPreLoopPhase() bool {
	return self.frameIndex < self.loopIndex
}
func (self *Animation) Restart() {
	self.frameIndex = 0
	self.frameDurationLeft = self.frameDurations[0]
}

var IdleAnimation Animation
var MoveAnimation Animation

// --- player ---

type Player struct {
	x, y float64
	animation *Animation
	direction int // -1 = left, 1 = right
	moving bool
}

func (self *Player) Update() {
	for range mipix.Tick().GetRate() {
		self.animation.Update()
		self.updateDirection()
		if self.moving {
			self.ensureAnimation(&MoveAnimation)
			preX := self.x
			switch self.animation.InPreLoopPhase() {
			case true  : self.x += float64(self.direction)*0.48
			case false : self.x += float64(self.direction)*1.24
			}
	
			self.x = min(max(self.x, 29.5), 284.5 - PlayerFrameWidth)
			if self.x == preX { self.moving = false }
		}
		if !self.moving {
			self.ensureAnimation(&IdleAnimation)
		}
	}
}

func (self *Player) DrawHiRes(target *ebiten.Image) {
	frame := self.animation.GetFrame()
	if self.direction == -1 {
		mipix.HiRes().DrawHorzFlip(target, frame, self.x, self.y)
	} else {
		mipix.HiRes().Draw(target, frame, self.x, self.y)
	}
}

func (self *Player) GetCameraCoords() (float64, float64) {
	return self.x + PlayerFrameWidth/2.0, self.y + PlayerFrameHeight/4.0
}

func (self *Player) updateDirection() {
	self.moving = true
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft)  : self.direction = -1
	case ebiten.IsKeyPressed(ebiten.KeyA)          : self.direction = -1
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight) : self.direction =  1
	case ebiten.IsKeyPressed(ebiten.KeyD)          : self.direction =  1
	default: self.moving = false
	}
}

func (self *Player) ensureAnimation(anim *Animation) {
	if self.animation == anim { return }
	self.animation = anim
	self.animation.Restart()
}

const PlayerFrameWidth  = 17
const PlayerFrameHeight = 51
func getPlayerFrameAt(spritesheet *ebiten.Image, row, col int) *ebiten.Image {
	rect := image.Rect(
		PlayerFrameWidth*col, PlayerFrameHeight*row,
		PlayerFrameWidth*(col + 1), PlayerFrameHeight*(row + 1),
	)
	return spritesheet.SubImage(rect).(*ebiten.Image)
}

// --- game ---

type Game struct {
	backGraphics  []Graphic
	frontGraphics []Graphic
	player Player
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

	// update zoom
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		_, targetZoom := mipix.Camera().GetZoom()
		if targetZoom == 1.0 {
			mipix.Camera().Zoom(1.5)
		} else {
			mipix.Camera().Zoom(1.0)
		}
	}

	// trigger shake
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		mipix.Camera().TriggerShake(0, 120, 60)
	}

	// update player and camera
	self.player.Update()
	x, y := self.player.GetCameraCoords()
	mipix.Camera().NotifyCoordinates(x, y)
	mipix.Redraw().Request()
	return nil
}

func (self *Game) Draw(canvas *ebiten.Image) {
	if !mipix.Redraw().Pending() { return }

	mipix.Debug().Drawf("[Q/E] %s filter", mipix.Scaling().GetFilter().String())
	mipix.Debug().Drawf("[A/D] Move")
	mipix.Debug().Drawf("[F] Fullscreen")
	mipix.Debug().Drawf("[Z] Zoom")
	mipix.Debug().Drawf("[S] Shake")

	canvas.Fill(color.RGBA{244, 232, 232, 255})
	self.DrawGraphics(canvas, self.backGraphics)
	mipix.QueueHiResDraw(self.DrawHiResPlayer)
	mipix.QueueDraw(self.DrawFrontGraphics)
}

func (self *Game) DrawGraphics(canvas *ebiten.Image, graphics []Graphic) {
	origin := mipix.Camera().Area().Min
	var opts ebiten.DrawImageOptions
	for _, graphic := range graphics {
		opts.GeoM.Translate(float64(graphic.X - origin.X), float64(graphic.Y - origin.Y))
		canvas.DrawImage(graphic.Source, &opts)
		opts.GeoM.Reset()
	}
}

func (self *Game) DrawHiResPlayer(viewport, target *ebiten.Image) {
	self.player.DrawHiRes(target)
}

func (self *Game) DrawFrontGraphics(canvas *ebiten.Image) {
	self.DrawGraphics(canvas, self.frontGraphics)
}

func main() {
	ebiten.SetWindowTitle("mipix-examples/src/gametest")
	mipix.SetResolution(GameWidth, GameHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(false)
	mipix.Redraw().SetManaged(true)

	// create animations
	pss := loadGraphic("player", 0, 0).Source // player spritesheet
	idle := getPlayerFrameAt(pss, 0, 0)
	tightFront1 := getPlayerFrameAt(pss, 0, 1)
	IdleAnimation.AddFrame(idle, 255)
	IdleAnimation.AddFrame(tightFront1, 30)
	IdleAnimation.AddFrame(idle, 30)
	IdleAnimation.AddFrame(getPlayerFrameAt(pss, 0, 3), 30)
	IdleAnimation.AddFrame(idle, 30)
	IdleAnimation.AddFrame(tightFront1, 30)
	IdleAnimation.AddFrame(idle, 80)
	
	t := uint8(8)
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 1, 0), 11)
	MoveAnimation.loopIndex = 1
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 2, 0), t)
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 2, 1), t)
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 2, 2), t)
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 2, 3), t)

	// set up everything for the game
	floorY := GameHeight - 34
	playerX, playerY := float64(106), float64(floorY - PlayerFrameHeight + 3)
	player := Player{ x: playerX, y: playerY, animation: &IdleAnimation, direction: 1 }
	game := &Game{
		backGraphics: []Graphic{
			loadGraphic("dark_floor_left_corner", 30, floorY),
			loadGraphic("dark_floor_side", 30, floorY + 21),
			loadGraphic("dark_floor_center", 67, floorY),
			loadGraphic("dark_floor_center", 127, floorY),
			loadGraphic("dark_floor_center", 187, floorY),
			loadGraphic("dark_floor_right_corner", 247, floorY),
			loadGraphic("dark_floor_side", 247, floorY + 21),
			loadGraphic("platform_flat_horz_small_A", -113, 60),
			loadGraphic("step_small_A", -45, 53),
			loadGraphic("step_small_B", -25, 46),
			loadGraphic("step_small_C", -5, 39),
			loadGraphic("step_small_A", 15, 32),
			loadGraphic("step_long_A", 35, 25),
			loadGraphic("step_small_D", 72, 18),
			loadGraphic("step_small_C", 92, 11),
			loadGraphic("step_long_A", 112, 4),
			loadGraphic("platform_flat_horz_small_A", 183, 26),
			loadGraphic("step_small_D", 251, 20),
			loadGraphic("platform_ground_square_small_B", -37, 93),
			loadGraphic("platform_ground_square_small_A", 310, 88),
			loadGraphic("large_sword_absorbed", 311, 15),
			loadGraphic("right_sign", -78, 33),
			loadGraphic("skeleton_A", 189, 19),
			loadGraphic("back_axe_A", 224, 83),
			loadGraphic("sword_D", -17, 65),
			loadGraphic("back_spear_A", 47, 55),
			loadGraphic("back_skull_A", -27, 87),
			loadGraphic("axe_A", 204, 3),
			loadGraphic("sword_A", 174, 82),
			loadGraphic("skull_B", 165, 102),
			loadGraphic("back_skeleton_A", 44, 103),
			loadGraphic("back_skull_A", 183, 104),
			loadGraphic("back_sword_B", 200, 69),
			loadGraphic("back_spear_B", 255, 55),
			loadGraphic("back_skull_B", 88, 104),
		},
		frontGraphics: []Graphic{
			loadGraphic("sword_B", 74, 69),
			loadGraphic("spear_A", 214, 55),
		},
		player: player,
	}

	// set camera initial position
	camX, camY := player.GetCameraCoords()
	mipix.Camera().ResetCoordinates(camX, camY)

	// run the game
	err := mipix.Run(game)
	if err != nil { panic(err) }
}
