package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"fmt"
	"os"

	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
	"github.com/voylento/learn-pub-sub-starter/internal/pubsub"
	"github.com/voylento/learn-pub-sub-starter/internal/routing"
)

func main() {
	const cs = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(cs)
	if err != nil {
		fmt.Printf("Error dialing rabbitmq connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Error: attempt to open channel failed: %v\n", err)
		os.Exit(1)
	}
	defer ch.Close()

	msg, err := gamelogic.ClientWelcome()
	fmt.Printf("User entered %s\n", msg)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	_, _, err = pubsub.DeclareAndBind(conn, "peril_direct", "pause." + msg, "pause", pubsub.Transient)
	if err != nil {
		fmt.Printf("Error calling DeclareAndBind: %v\n", err)
		os.Exit(1)
	}

	gs := gamelogic.NewGameState(msg)
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, "pause." + msg, routing.PauseKey, pubsub.Transient, handlerPause(gs))
	if err != nil {
		fmt.Printf("Error subscribing to pause notifications: %v\n", err)  
		os.Exit(1)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "army_moves." + msg, "army_moves.*", pubsub.Transient, handlerArmyMoves(gs))
	if err != nil {
		fmt.Printf("Error subscribing to army move notifications: %v\n", err)
		os.Exit(1)
	}

	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			fmt.Println("Error: no input received")
			continue
		}

		switch userInput[0] {
		case "spawn":
			if err = gs.CommandSpawn(userInput); err != nil {
				fmt.Println(err)
			}
		case "move":
			mv, err := gs.CommandMove(userInput) 
			if err != nil {
				fmt.Println(err)
			}
			pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				"army_moves." + msg,
				mv)
			fmt.Println("move published successfully")
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

