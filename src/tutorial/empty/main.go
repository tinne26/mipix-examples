package main

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/tinne26/mipix"

type Game struct {}

func (game *Game) Update() error {
	return nil
}

func (game *Game) Draw(canvas *ebiten.Image) {
	// ...
}

func main() {
	mipix.SetResolution(128, 72)
	err := mipix.Run(&Game{})
	if err != nil { panic(err) }
}
