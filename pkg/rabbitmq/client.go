package rabbitmq

import (
	"context"
	"errors"
	"github.com/streadway/amqp"
	"github.com/teneta-io/dcp/internal/config"
	"go.uber.org/zap"
	"strings"
	"sync/atomic"
	"time"
)

var (
	ErrAMQPConfigurationEmpty              = errors.New("amqp url is empty")
	ErrConnectionRecreateFailed            = errors.New("connection recreate failed")
	ErrChannelRecreateFailed               = errors.New("channel recreate failed")
	ErrCouldNotEstablishRabbitMQConnection = errors.New("could not establish rabbitMQ connection")
	ErrCouldNotOpenConnectionChannel       = errors.New("could not open connection channel")
	ErrSendBeforeEstablishConnection       = errors.New("tried to send message before connection was initialized")
)

type Message struct {
	body      []byte
	queueName string
}

type RabbitMQ struct {
	ctx        context.Context
	logger     *zap.Logger
	connection *Connection
}

type Connection struct {
	logger *zap.Logger
	*amqp.Connection
}

type Channel struct {
	logger *zap.Logger
	*amqp.Channel
	closed int32
}

func NewClient(ctx context.Context, cfg *config.RabbitMQConfig, logger *zap.Logger) (*RabbitMQ, error) {
	var err error
	client := &RabbitMQ{ctx: ctx, logger: logger.Named("RabbitMQClient")}
	dsnList := strings.Split(cfg.DSNList, ";")

	if len(dsnList) < 1 {
		client.logger.Error(ErrAMQPConfigurationEmpty.Error())
		return nil, ErrAMQPConfigurationEmpty
	}

	if err = client.Connect(dsnList); err != nil {
		return nil, err
	}

	return client, nil
}

func (client *RabbitMQ) Connect(dsnList []string) error {
	client.logger.Info("Connecting to rabbitMQ...")

	nodeSequence := 0
	conn, err := amqp.DialConfig(dsnList[nodeSequence], amqp.Config{
		Heartbeat: 5 * time.Second,
	})

	if err != nil {
		client.logger.Error(ErrCouldNotEstablishRabbitMQConnection.Error(), zap.Any("error", err))
		return err
	}

	client.connection = &Connection{
		client.logger,
		conn,
	}

	go func(urls []string, seq *int) {
		for {
			reason, ok := <-client.connection.NotifyClose(make(chan *amqp.Error))

			if !ok {
				client.logger.Debug("connection closed")
				break
			}

			client.logger.Info("connection closed", zap.Any("reason", reason))

			for {
				time.Sleep(3 * time.Second)
				newSeq := client.next(urls, *seq)
				*seq = newSeq

				conn, err := amqp.Dial(urls[newSeq])

				if err == nil {
					client.connection.Connection = conn
					break
				}

				client.logger.Error(ErrConnectionRecreateFailed.Error(), zap.Any("error", err))
			}
		}
	}(dsnList, &nodeSequence)

	return nil
}

func (c *Connection) Channel() (*Channel, error) {
	ch, err := c.Connection.Channel()

	if err != nil {
		return nil, err
	}

	channel := &Channel{
		Channel: ch,
		logger:  c.logger,
	}

	go func() {
		for {
			reason, ok := <-channel.NotifyClose(make(chan *amqp.Error))

			if !ok {
				c.logger.Debug("channel closed")
				break
			}

			c.logger.Info("channel closed", zap.Any("reason", reason))

			for {
				time.Sleep(3 * time.Second)
				ch, err := c.Connection.Channel()

				if err == nil {
					channel.Channel = ch
					break
				}

				c.logger.Error(ErrChannelRecreateFailed.Error(), zap.Any("error", err))
			}
		}
	}()

	return channel, nil
}

func (ch *Channel) IsClosed() bool {
	return atomic.LoadInt32(&ch.closed) == 1
}

func (ch *Channel) Close() error {
	if ch.IsClosed() {
		return amqp.ErrClosed
	}

	atomic.StoreInt32(&ch.closed, 1)

	return ch.Channel.Close()
}

func (client *RabbitMQ) Close() error {
	if client.connection != nil {
		return client.connection.Close()
	}

	return nil
}

func (ch *Channel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	deliveries := make(chan amqp.Delivery)

	go func() {
		for {
			del, err := ch.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)

			if err != nil {
				time.Sleep(3 * time.Second)
				continue
			}

			for msg := range del {
				deliveries <- msg
			}

			time.Sleep(3 * time.Second)

			if ch.IsClosed() {
				break
			}
		}
	}()

	return deliveries, nil
}

func (client *RabbitMQ) failOnError(msg string, err error) {
	if err != nil {
		client.logger.Panic(msg, zap.Any("error", err))
	}
}

func (client *RabbitMQ) next(s []string, lastSeq int) int {
	length := len(s)
	if length == 0 || lastSeq == length-1 {
		return 0
	} else if lastSeq < length-1 {
		return lastSeq + 1
	} else {
		return -1
	}
}

func (client *RabbitMQ) consume(deliveries <-chan amqp.Delivery, handlerFunc func(d amqp.Delivery)) {
	for {
		select {
		case delivery := <-deliveries:
			handlerFunc(delivery)
		case <-client.ctx.Done():
			return
		}
	}
}
