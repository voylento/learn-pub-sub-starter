package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
	"github.com/voylento/learn-pub-sub-starter/internal/pubsub"
	"github.com/voylento/learn-pub-sub-starter/internal/routing"
)

func handlerGameLog(conn *amqp.Connection) func(routing.GameLog) pubsub.AckType {
	return func(gameLog routing.GameLog) pubsub.AckType {
		fmt.Printf("Received GameLog message from queue: %v\n", gameLog)
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gameLog)
		if err != nil {
			fmt.Printf("GameLog message received, but error writing it to disk: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
