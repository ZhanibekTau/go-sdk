package rabbitmq

import (
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"go-sdk/pkg/rabbitmq/enums"
)

type Config func(*amqp.Config)

// WithAmqpURI - конфиг лоя лобавления ссылки коннект
func WithAmqpURI(uri string) Config {
	return func(config *amqp.Config) {
		config.Connection.AmqpURI = uri
	}
}

// WithMarshaller - конфиг для маршалинга сообщений
func WithMarshaller(m amqp.DefaultMarshaler) Config {
	return func(config *amqp.Config) {
		config.Marshaler = m
	}
}

// WithExchangeConfig - add exchange ExchangeConfig
func WithExchangeConfig(ec amqp.ExchangeConfig) Config {
	return func(config *amqp.Config) {
		config.Exchange = ec
	}
}

// WithPublishConfig - add publish config PublishConfig
func WithPublishConfig(pc amqp.PublishConfig) Config {
	return func(config *amqp.Config) {
		config.Publish = pc
	}
}

// WithConsumerConfig - для добавление конфига консьюмера
func WithConsumerConfig(cc amqp.ConsumeConfig) Config {
	return func(config *amqp.Config) {
		config.Consume = cc
	}
}

// WithRoutingKeyBinding - добавление конфига для бинда с routing key
func WithRoutingKeyBinding(rk amqp.QueueBindConfig) Config {
	return func(config *amqp.Config) {
		config.QueueBind = rk
	}
}

// WithQueueName  - добавление конфгиа для названия очереди
func WithQueueName(q amqp.QueueConfig) Config {
	return func(config *amqp.Config) {
		config.Queue = q
	}
}

// WithTopologyBuilder
func WithTopologyBuilder(tp amqp.TopologyBuilder) Config {
	if tp == nil {
		tp = &amqp.DefaultTopologyBuilder{}
	}

	return func(config *amqp.Config) {
		config.TopologyBuilder = tp
	}
}

// NewConsumerTopicDurableConfig - дефолтный конфиг для консьюмера
func NewConsumerTopicDurableConfig(routingKey, exchange, queue string, prefetchCnt int) []Config {
	return []Config{
		WithMarshaller(amqp.DefaultMarshaler{NotPersistentDeliveryMode: true}),
		WithRoutingKeyBinding(amqp.QueueBindConfig{
			GenerateRoutingKey: func(topic string) string { return routingKey },
		}),
		WithConsumerConfig(amqp.ConsumeConfig{
			Qos: amqp.QosConfig{PrefetchCount: prefetchCnt},
		}),
		WithExchangeConfig(amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return exchange
			},
			Type:    enums.TopicExchange,
			Durable: true,
		}),
		WithQueueName(amqp.QueueConfig{
			GenerateName: func(topic string) string {
				return queue
			},
			Durable: true,
		}),
		WithTopologyBuilder(&amqp.DefaultTopologyBuilder{}),
	}
}

// NewPublisherTopicDurableConfig - дефолтный конфиг для publisher
func NewPublisherTopicDurableConfig(routingKey, exchange string) []Config {
	return []Config{
		WithMarshaller(amqp.DefaultMarshaler{NotPersistentDeliveryMode: true}),
		WithPublishConfig(amqp.PublishConfig{
			GenerateRoutingKey: func(topic string) string { return routingKey },
			ConfirmDelivery:    true,
		}),
		WithRoutingKeyBinding(amqp.QueueBindConfig{
			GenerateRoutingKey: func(topic string) string { return routingKey },
		}),
		WithExchangeConfig(amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return exchange
			},
			Type:    enums.TopicExchange,
			Durable: true,
		}),
		WithTopologyBuilder(&amqp.DefaultTopologyBuilder{}),
	}
}
