package main

import (
	"encoding/json"

	"github.com/Dsmit05/img-loader/internal/logger"
	"github.com/pkg/errors"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type Producer struct {
	p     *kafka.Producer
	topic string
}

func NewProducer(uri, topic string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": uri})
	if err != nil {
		return nil, errors.Wrap(err, "kafka.NewProducer() err")
	}

	return &Producer{p: p, topic: topic}, nil
}

func (p *Producer) Start() {
	go func() {
		for e := range p.p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logger.Error("Delivery failed", ev.TopicPartition.Error)
				} else {
					logger.Info("Delivered", "Delivered message to TopicPartition")
				}
			}
		}
	}()
}

type UserEvent struct {
	FirstName string
	LastName  string
	URL       string
}

func (p *Producer) Send(img UserEvent) error {
	v, err := json.Marshal(img)
	if err != nil {
		return errors.Wrap(err, "json.Marshal() error")
	}

	err = p.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          v,
	}, nil)
	if err != nil {
		return errors.Wrap(err, "p.Produce() error")
	}

	p.p.Flush(15 * 10)
	return nil
}

func (p *Producer) Stop() {
	p.p.Close()
}
