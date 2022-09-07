package rmq

import (
	"encoding/json"
	"time"

	"github.com/Dsmit05/img-loader/internal/logger"
	"github.com/Dsmit05/img-loader/internal/models"
	"github.com/Dsmit05/img-loader/pkg/rmq"
	"github.com/pkg/errors"
)

const consumerTag = "img-loader"

type Consumer struct {
	rmq   *rmq.Consumer
	chImg chan models.UserEvent
}

func NewConsumer(uri, queueName string) (*Consumer, error) {
	consumerConfig := &rmq.ConsumerConfig{
		Name:                  consumerTag,
		ReconnectTimeout:      time.Second * 5,
		RetryReconnectCount:   5,
		DeliveryChannBuffSize: 10,
		URI:                   uri,
		PrefetchCount:         5,
		PrefetchSize:          0,
		Queue:                 queueName,
	}

	consumer := rmq.NewConsumer(consumerConfig)
	err := consumer.InitConsumer()
	if err != nil {
		return nil, errors.Wrap(err, " rmq.NewConsumer() err")
	}

	return &Consumer{
		rmq:   consumer,
		chImg: make(chan models.UserEvent, 10),
	}, nil
}

func (c *Consumer) Process() {
	c.rmq.StartConsume()

	for {
		select {
		case delivery, ok := <-c.rmq.DeliveryChannel:
			if !ok {
				logger.Info("c.rmq.DeliveryChannel", "chan close")
				return
			}

			var userEvent models.UserEvent
			err := json.Unmarshal(delivery.Body, &userEvent)
			if err != nil {
				logger.Error("deserialize rmq msg error", err)
				if err := c.rmq.Nack(delivery, false, false); err != nil {
					logger.Error(" c.rmq.Nack() error", err)
				}
				continue
			}

			if err := c.rmq.Ack(delivery, false); err != nil {
				logger.Error(" c.rmq.Ack() error", err)
			}

			c.chImg <- userEvent

		}
	}
}

func (c *Consumer) GetChanIMG() <-chan models.UserEvent {
	return c.chImg
}

func (c *Consumer) Stop() {
	c.rmq.Stop()
	close(c.chImg)
}
