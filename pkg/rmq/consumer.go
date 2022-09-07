package rmq

import (
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/streadway/amqp"
)

const (
	flushingTimeout = 100 * time.Millisecond

	notActiveConsumerTimeout = time.Second * 60 * 20 // 20 min
)

type Consumer struct {
	DeliveryChannel chan amqp.Delivery
	config          *ConsumerConfig

	stopConsumingFlag bool // ToDo atomic

	connection      *amqp.Connection
	channel         *amqp.Channel
	deliveryChannel <-chan amqp.Delivery

	workerWaitGroup *sync.WaitGroup

	reconnectionMutex *sync.Mutex
	reconnectionFlag  bool // ToDo atomic
	stopMust          bool // ToDo atomic
}

func NewConsumer(config *ConsumerConfig) *Consumer {
	workerWaitGroup := sync.WaitGroup{}
	deliveryBufferChannel := make(chan amqp.Delivery, config.DeliveryChannBuffSize)

	consumer := &Consumer{
		DeliveryChannel:   deliveryBufferChannel,
		config:            config,
		workerWaitGroup:   &workerWaitGroup,
		reconnectionMutex: &sync.Mutex{},
	}
	return consumer
}

func (c *Consumer) InitConsumer() error {
	connection, err := amqp.Dial(c.config.URI)
	if err != nil {
		return errors.Wrap(err, "amqp.Dial() error")
	}

	c.connection = connection
	channel, err := c.connection.Channel()
	if err != nil {
		c.connection.Close()
		return errors.Wrap(err, "c.connection.Channel() error")
	}

	c.channel = channel

	err = c.channel.Qos(c.config.PrefetchCount, c.config.PrefetchSize, c.config.PrefetchGlobal)
	if err != nil {
		c.channel.Close()
		c.connection.Close()
		return errors.Wrap(err, "c.channel.Qos() error")
	}

	deliveryChannel, err := c.channel.Consume(
		c.config.Queue,
		c.config.Name,
		c.config.AutoAck,
		c.config.Exclusive,
		c.config.NoLocal,
		c.config.NoWait,
		nil,
	)

	if err != nil {
		c.channel.Close()
		c.connection.Close()
		return errors.Wrap(err, "c.channel.Consume() error")
	}

	c.deliveryChannel = deliveryChannel

	return nil
}

func (c *Consumer) StartConsume() {
	go c.consume()
}

func (c *Consumer) consume() {
	ticker := time.NewTicker(notActiveConsumerTimeout)

	c.workerWaitGroup.Add(1)
	defer func() {
		ticker.Stop()
		c.workerWaitGroup.Done()
	}()

	var ok bool
	var delivery amqp.Delivery

	for {
		if c.stopConsumingFlag {
			return
		}
		select {
		case delivery, ok = <-c.deliveryChannel:
			if !ok {
				go c.reconnect()
				return
			}
			c.DeliveryChannel <- delivery
		case <-ticker.C:
			_, err := amqp.Dial(c.config.URI)
			ticker.Reset(notActiveConsumerTimeout)
			if err != nil {
				go c.reconnect()
			}

		}
	}
}

func (c *Consumer) reconnect() {
	if c.reconnectionFlag {
		return
	}

	c.reconnectionMutex.Lock()
	c.reconnectionFlag = true
	defer c.reconnectionMutex.Unlock()

	// retry reconnect
	var n uint
	for {
		c.stopConsume()
		if c.channel != nil {
			c.channel.Close()
		}
		if c.connection != nil {
			c.connection.Close()
		}

		time.Sleep(c.config.ReconnectTimeout)

		err := c.InitConsumer()
		if err != nil {
			n++
			if n == c.config.RetryReconnectCount { // todo: atomic
				c.Stop()
				return
			}
			continue
		}

		c.reconnectionFlag = false
		c.StartConsume()

		return
	}
}

func (c *Consumer) cleanPublicDelivery() {
	var emptiCheck bool
	for {
		select {
		case <-c.DeliveryChannel:
			emptiCheck = false
		default:
			if !emptiCheck {
				time.Sleep(flushingTimeout)
				emptiCheck = true
				continue
			}
			return
		}
	}
}

func (c *Consumer) Ack(delivery amqp.Delivery, multiple bool) error {
	if c.reconnectionFlag {
		return nil
	}

	if c.stopMust {
		return errors.New("rmq consumer STOP")
	}

	err := delivery.Ack(multiple)
	if err != nil {
		go c.reconnect()
	}

	return nil
}

func (c *Consumer) Nack(delivery amqp.Delivery, multiple bool, requeue bool) error {
	if c.reconnectionFlag {
		return nil
	}

	if c.stopMust {
		return errors.New("rmq consumer STOP")
	}

	err := delivery.Nack(multiple, requeue)
	if err != nil {
		go c.reconnect()
	}

	return nil
}

func (c *Consumer) stopConsume() {
	c.stopConsumingFlag = true
	c.cleanPublicDelivery()
	c.workerWaitGroup.Wait()
	c.stopConsumingFlag = false

	if c.channel != nil {
		c.channel.Close()
	}
	if c.connection != nil {
		c.connection.Close()
	}
}

func (c *Consumer) Stop() {
	c.stopMust = true
	c.stopConsumingFlag = true
	c.cleanPublicDelivery()

	if c.channel != nil {
		c.channel.Close()
	}
	if c.connection != nil {
		c.connection.Close()
	}

	close(c.DeliveryChannel)
}
