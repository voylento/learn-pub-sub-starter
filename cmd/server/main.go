package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"

	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
	ps "github.com/voylento/learn-pub-sub-starter/internal/pubsub"
	"github.com/voylento/learn-pub-sub-starter/internal/routing"
)

// learn-pub-sub-starter/cmd/server/main.go
func main() {
	const cs = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(cs)
	if err != nil {
		fmt.Printf("Error dialing rabbitmq connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connection successful. Starting Peril server...")

	// Create queue to notify clients when game is paused
	// Note: consider creating just the exchange in server and
	// creating the channel in the client
	_, _, err = ps.DeclareAndBind(
		conn,                       // *amqp.Connection
		routing.ExchangePerilTopic, // exchange
		routing.GameLogSlug,        // queue name
		routing.GameLogSlug+".*",   // key
		ps.Durable)
	if err != nil {
		fmt.Printf("Error calling DeclareAndBind: %v\n", err)
		os.Exit(1)
	}

	subscribeToGameLogMessages(conn)

	gamelogic.PrintServerHelp()

	// get channel for use when publishing pause message
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Error: attemtp to open channel failed: %v\n", err)
		os.Exit(1)
	}
	defer ch.Close()

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

func subscribeToGameLogMessages(conn *amqp.Connection) {
	_, _, err := ps.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		ps.Durable,
	)
	if err != nil {
		fmt.Printf("Error creating queue for game logs. Game logs will not be available: %v\n", err)
		return
	}

	err = ps.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		ps.Durable,
		handlerGameLog(conn),
	)
	if err != nil {
		fmt.Printf("Error subscribing to game log messages: %v\n", err)
	}
}
