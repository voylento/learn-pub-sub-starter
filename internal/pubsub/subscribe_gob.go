package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func SubscribeGob[T any](
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

	ch.Qos(10, 0, false)

	delivery_ch, err := ch.Consume(queueName, "", false, false, false, false, nil)

	go func(ch <-chan amqp.Delivery) {
		for delivery := range ch {
			var payload T
			dec := gob.NewDecoder(bytes.NewReader(delivery.Body))
			if err = dec.Decode(&payload); err != nil {
				log.Printf("Failed to unmashal gob: %v", err)
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
