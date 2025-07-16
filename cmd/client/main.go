package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"strconv"
	"time"

	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
	"github.com/voylento/learn-pub-sub-starter/internal/pubsub"
	"github.com/voylento/learn-pub-sub-starter/internal/routing"
)

// learn-pub-sub-starter/cmd/client/main.go
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

	userName, err := gamelogic.ClientWelcome()
	fmt.Printf("User entered %s\n", userName)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	gs := gamelogic.NewGameState(userName)

	ch, _, _ = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug,
		pubsub.Durable)

	subscribeToPauseMessages(conn, gs, userName)
	subscribeToArmyMoveMessages(conn, gs, userName)
	subscribeToWarMessages(conn, gs)

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
				"army_moves."+userName,
				mv)
			fmt.Println("move published successfully")
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(userInput) < 2 {
				fmt.Println("Usage: spam number")
				return
			}
			n, err := strconv.Atoi(userInput[1])
			if err != nil {
				fmt.Printf("Error converting %s to integer: %v\n", userInput[1], err)
				return
			}
			spamGameQueue(ch, n, userName)
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func spamGameQueue(ch *amqp.Channel, n int, userName string) {

	for _ = range n {
		msg := gamelogic.GetMaliciousLog()

		log := routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    userName,
		}

		pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+userName, log)
	}
}

func subscribeToPauseMessages(conn *amqp.Connection, gs *gamelogic.GameState, userName string) {
	pauseQueue := routing.PauseKey + "." + userName

	_, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect, // exchange
		pauseQueue,                  // queue name
		routing.PauseKey,            // key
		pubsub.Transient,            // queue type (transient or durable)
	)
	if err != nil {
		fmt.Printf("Error calling DeclareAndBind: %v\n", err)
		os.Exit(1)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		pauseQueue,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(conn, gs))
	if err != nil {
		fmt.Printf("Error subscribing to pause notifications: %v\n", err)
		os.Exit(1)
	}

}

func subscribeToArmyMoveMessages(conn *amqp.Connection, gs *gamelogic.GameState, userName string) {
	armyMovesQueue := routing.ArmyMovesPrefix + "." + userName
	armyMovesKey := routing.ArmyMovesPrefix + ".*"
	err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,           // exchange name
		armyMovesQueue,                       // queue name
		armyMovesKey,                         // key
		pubsub.Transient,                     // queue type
		handlerArmyMoves(conn, gs, userName), // handler
	)
	if err != nil {
		fmt.Printf("Error subscribing to army move notifications: %v\n", err)
		os.Exit(1)
	}
}

func subscribeToWarMessages(conn *amqp.Connection, gs *gamelogic.GameState) {
	warQueue := routing.WarRecognitionsPrefix
	warKey := routing.WarRecognitionsPrefix + ".*"
	err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic, // exchange name
		warQueue,                   // queue name
		warKey,                     // key
		pubsub.Durable,             // queue type
		handlerWar(conn, gs),
	)
	if err != nil {
		fmt.Printf("Error subscribing to war notifications: %v\n", err)
		os.Exit(1)
	}
}
