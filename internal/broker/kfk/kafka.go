package kfk

import (
	"encoding/json"

	"github.com/Dsmit05/img-loader/internal/logger"
	"github.com/Dsmit05/img-loader/internal/models"
	"github.com/pkg/errors"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type Consumer struct {
	consumer *kafka.Consumer
	chImg    chan models.UserEvent
}

func NewConsumer(uri, group, topic string) (*Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": uri,
		"group.id":          group,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, errors.Wrap(err, "kafka.NewConsumer() err")
	}

	if err := consumer.SubscribeTopics([]string{topic}, nil); err != nil {
		return nil, errors.Wrap(err, "consumer.SubscribeTopics() err")
	}

	return &Consumer{consumer: consumer, chImg: make(chan models.UserEvent, 10)}, nil
}

func (c *Consumer) Process() {
	var Msg models.UserEvent

	for {
		msg, err := c.consumer.ReadMessage(-1)
		if err == nil {
			if err := json.Unmarshal(msg.Value, &Msg); err != nil {
				logger.Error("json.Unmarshal()", err)
				continue
			}

			c.chImg <- Msg

		} else {
			logger.Error("c.consumer.ReadMessage(-1)", err)
		}
	}
}

func (c *Consumer) GetChanIMG() <-chan models.UserEvent {
	return c.chImg
}

func (c *Consumer) Stop() {
	c.consumer.Close()
	close(c.chImg)
}
