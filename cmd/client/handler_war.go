package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
	"github.com/voylento/learn-pub-sub-starter/internal/pubsub"
	"github.com/voylento/learn-pub-sub-starter/internal/routing"
	"time"
)

func handlerWar(conn *amqp.Connection, gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(warMove gamelogic.RecognitionOfWar) pubsub.AckType {
		fmt.Println("Inside War Handler")
		fmt.Printf("warMove.Attacker.Name = %s, warMove.Defender.Name = %s\n", warMove.Attacker.Username, warMove.Defender.Username)
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(warMove)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeYouWon:
			logMsg := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := publishGameLog(conn, winner, logMsg)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			logMsg := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := publishGameLog(conn, winner, logMsg)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			logMsg := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := publishGameLog(conn, winner, logMsg)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("Unknown war outcome. NackDiscard")
			return pubsub.NackDiscard
		}
	}
}

func publishGameLog(conn *amqp.Connection, attacker, message string) error {
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     message,
		Username:    attacker,
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+attacker, log)
	return err
}
