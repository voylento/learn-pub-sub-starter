package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
	"github.com/voylento/learn-pub-sub-starter/internal/pubsub"
	"github.com/voylento/learn-pub-sub-starter/internal/routing"
)

func handlerArmyMoves(conn *amqp.Connection, gs *gamelogic.GameState, user_name string) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(armyMove gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(armyMove)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			fmt.Println("MoveOutcomeSafe, Ack")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			fmt.Println("Move Handler: MAKE WAR!!!")
			defender := gs.GetPlayerSnap()
			fmt.Printf("Attacker: %s\n", armyMove.Player.Username)
			fmt.Printf("Defender: %s\n", defender.Username)
			warObj := gamelogic.RecognitionOfWar{
				Attacker: armyMove.Player,
				Defender: gs.GetPlayerSnap(),
			}
			err := publishWarMessage(
				conn,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+user_name,
				warObj)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			fmt.Println("MoveOutcomeSamePlayer, NackDiscard")
			return pubsub.NackDiscard
		default:
			fmt.Println("Unknown outcome. NackDiscard")
			return pubsub.NackDiscard
		}
	}
}

func publishWarMessage(conn *amqp.Connection, exchange, routing_key string, warObj gamelogic.RecognitionOfWar) error {
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Error creating channel for publishing MakeWar message: %v\n", err)
	}
	defer ch.Close()

	err = pubsub.PublishJSON(
		ch,
		exchange,
		routing_key,
		warObj)
	return err
}
