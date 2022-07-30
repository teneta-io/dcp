package rabbitmq

import "github.com/streadway/amqp"

const (
	queueName   = "tasks"
	consumeName = "dcp"
)

type TaskSubscriber struct {
	client       *RabbitMQ
	consumerName string
	queueName    string
	args         amqp.Table
	ch           *Channel
}

func NewTaskSubscriber(client *RabbitMQ) *TaskSubscriber {
	return &TaskSubscriber{
		client:       client,
		consumerName: consumeName,
		queueName:    queueName,
	}
}

func (subscriber *TaskSubscriber) Subscribe(handlerFunc func(delivery amqp.Delivery)) error {
	ch, err := subscriber.client.connection.Channel()
	defer ch.Close()

	subscriber.client.failOnError("Failed to open a channel", err)

	queue, err := ch.QueueDeclare(
		subscriber.queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	subscriber.client.failOnError("Failed to register an Queue", err)

	msgs, err := ch.Consume(
		queue.Name,
		subscriber.consumerName,
		true,
		false,
		false,
		false,
		nil,
	)

	subscriber.client.failOnError("Failed to register a consumer", err)
	subscriber.client.consume(msgs, handlerFunc)

	return nil
}
