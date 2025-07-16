package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"fmt"

	"github.com/voylento/learn-pub-sub-starter/internal/gamelogic"
	"github.com/voylento/learn-pub-sub-starter/internal/routing"
	"github.com/voylento/learn-pub-sub-starter/internal/pubsub"
)

func handlerPause(_ *amqp.Connection, gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType{
	return func(ps routing.PlayingState) pubsub.AckType {
		fmt.Println("Inside pause handler")
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		fmt.Println("returning pubsub.Ack")
		return pubsub.Ack
	}
}

