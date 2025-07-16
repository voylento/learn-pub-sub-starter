package pubsub

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Error: attemtp to open channel failed: %v\n", err)
		return nil, amqp.Queue{}, err
	}

	isTransient := queueType == Transient

	q, err := ch.QueueDeclare(
		queueName,
		!isTransient,
		isTransient,
		isTransient,
		false,
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
	if err != nil {
		fmt.Printf("Error creating queue: %v\n", err)
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Printf("Error binding queue: %v\n", err)
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}
