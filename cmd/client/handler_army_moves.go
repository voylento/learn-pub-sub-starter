package main

import (
	"fmt"

	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
)

func handlerArmyMoves(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(armyMove gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		fmt.Print("Successfully detected Army  Move")
		gs.HandleMove(armyMove)
	}
}

