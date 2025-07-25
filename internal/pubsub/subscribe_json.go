package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int
const (
	Ack	AckType = iota
	NackRequeue 
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {

	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	delivery_ch, err := ch.Consume(queueName, "", false, false, false, false, nil)

	go func(ch <-chan amqp.Delivery) {
		for delivery := range ch {
			var payload T

			// Unmarshal the JSON bytes into the generic type T
			if err := json.Unmarshal(delivery.Body, &payload); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				delivery.Nack(false, false)
				continue
			}
			ack := handler(payload)
			switch ack {
			case Ack:
				fmt.Println("Ack")
				delivery.Ack(false)
			case NackRequeue:
				fmt.Println("NackRequeue")
				delivery.Nack(false, true)
			case NackDiscard:
				fmt.Println("NackDiscard")
				delivery.Nack(false, false)
			default:
				fmt.Printf("unknown ack return: %d\n", ack)
			}
		}
	}(delivery_ch)

	return nil
}
