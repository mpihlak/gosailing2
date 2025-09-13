package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mpihlak/ebiten-sailing/pkg/game"
)

func main() {
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Ebiten Sailing")

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
