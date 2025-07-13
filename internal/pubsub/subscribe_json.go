package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T),
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
			handler(payload)
			delivery.Ack(false)
		}
	}(delivery_ch)

	return nil
}
