package rmq

import (
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/streadway/amqp"
)

const (
	reconnectTimeout = 5 * time.Second
	maxTryPublish    = 2
	reconnectsMax    = 2
)

type Producer struct {
	URI string

	chann *amqp.Channel
	conn  *amqp.Connection

	sync.Mutex
}

func NewProducer(rabbitmqURI string) *Producer {
	rmqPublisher := &Producer{
		URI: rabbitmqURI,
	}
	return rmqPublisher
}

func (p *Producer) InitChannel() error {
	conn, err := amqp.Dial(p.URI)
	if err != nil {
		return errors.Wrap(err, "amqp.Dial() error")
	}

	p.conn = conn
	channel, err := p.conn.Channel()
	if err != nil {
		return errors.Wrap(err, " p.conn.Channel() error")
	}
	p.chann = channel
	return nil
}

func (p *Producer) Reconnect() error {
	p.Lock()
	defer p.Unlock()

	var n int
	for {
		n++
		err := p.InitChannel()
		if err != nil {
			if n == reconnectsMax {
				return errors.New("reconnectsMax limit, stop Reconnect")
			}
			time.Sleep(reconnectTimeout)
			continue
		}

		break
	}

	return nil
}

// Publish message to RabbitMQ exchange
func (p *Producer) publishToChannel(exchangeName, key string, msg []byte) error {
	message := amqp.Publishing{
		Headers:         amqp.Table{},
		ContentType:     "application/json",
		ContentEncoding: "",
		Body:            msg,
		DeliveryMode:    amqp.Persistent,
	}
	err := p.chann.Publish(
		exchangeName,
		key,
		false,
		false,
		message,
	)
	if err != nil {
		return errors.Wrap(err, "p.chann.Publish() err")
	}

	return nil
}

// Send message to RabbitMQ exchange
func (p *Producer) processMessage(exchangeName, key string, messageJSON []byte) error {
	var n int
	for {
		n++
		err := p.publishToChannel(exchangeName, key, messageJSON)
		if err != nil {
			errReconnect := p.Reconnect()
			if errReconnect != nil && n == maxTryPublish {
				return errors.Wrap(err, "errReconnect")
			}
			continue
		}
		break
	}

	return nil
}

func (p *Producer) PublishBytes(exchangeName, key string, data []byte) error {
	return p.processMessage(exchangeName, key, data)
}

func (p *Producer) CloseChannel() {
	if p.chann != nil {
		p.chann.Close()
		p.chann = nil
	}

	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
}
