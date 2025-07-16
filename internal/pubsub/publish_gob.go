package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"

	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(val)
	if err != nil {
		return fmt.Errorf("failed to marshal val to gob in PublishGob: %w", err)
	}

	return ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        data.Bytes(),
		},
	)
}
