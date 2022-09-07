package rmq

import "time"

type ConsumerConfig struct {
	Name                  string
	ReconnectTimeout      time.Duration
	RetryReconnectCount   uint
	DeliveryChannBuffSize int

	URI string

	PrefetchCount  int
	PrefetchSize   int
	PrefetchGlobal bool

	Queue     string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
}
