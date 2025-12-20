package kafka

import "time"

type ProducerConfig struct {
	Brokers     []string
	MaxMessages int
	Frequency   time.Duration
}

type ConsumerGroupConfig struct {
	Brokers             []string
	Topics              []string
	SleepAfterRebalance time.Duration
	GroupID             string
	AutoCommit          bool
}
