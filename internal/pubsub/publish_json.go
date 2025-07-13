package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"context"
	"encoding/json"
	"fmt"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal val to json in PublishJSON: %w", err)
	}

	return ch.PublishWithContext(
		context.Background(), 
		exchange,
		key,
		false,	// mandatory
		false,	// immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:				 jsonData,
		},
	)
}

