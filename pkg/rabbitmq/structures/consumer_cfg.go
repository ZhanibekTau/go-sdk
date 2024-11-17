package structures

import (
	"context"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Handler func(ctx context.Context, msg *message.Message) error

type ConsumeConfig struct {
	Handler  Handler
	Consumer *amqp.Subscriber
}
