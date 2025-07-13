package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"fmt"
	"os"

	ps "github.com/voylento/learn-pub-sub-starter/internal/pubsub"
	"github.com/voylento/learn-pub-sub-starter/internal/routing"
	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
)

func main() {
	const cs = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(cs)
	if err != nil {
		fmt.Printf("Error dialing rabbitmq connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connection successful. Starting Peril server...")

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Error: attemtp to open channel failed: %v\n", err)
		os.Exit(1)
	}
	defer ch.Close()

	_, _, err = ps.DeclareAndBind(conn, "peril_topic", "game_logs", "game_logs.*", ps.Durable)
	if err != nil {
		fmt.Printf("Error calling DeclareAndBind: %v\n", err)
		os.Exit(1)
	}

	gamelogic.PrintServerHelp()
	
	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			fmt.Println("Error: no input received")
			continue
		}

		switch userInput[0] {
		case "pause":
			fmt.Println("sending pause notification to all clients...")
			err = ps.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				})
			if err != nil {
				fmt.Printf("Error attempting to send pause command...%v\n", err)
			}
		case "resume":
			fmt.Println("sending resume notification to all clients...")
			err = ps.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				})
			if err != nil {
				fmt.Printf("Error sending resume command...%v\n", err)
			}
		case "help":
			gamelogic.PrintServerHelp()
		case "quit":
			fmt.Println("quitting server...")
			return
		default:
			fmt.Printf("Unknown command: %s\n", userInput[0])
			gamelogic.PrintServerHelp()
		}
	}
}
