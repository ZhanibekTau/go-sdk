package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"go-sdk/pkg/rabbitmq/structures"
	"go-sdk/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
	"log"
	"reflect"
	"runtime"
	"sync"
)

// AmqpPubSub
type AmqpPubSub struct {
	config   *amqp.Config
	conn     *amqp.ConnectionWrapper
	handlers []structures.ConsumeConfig
}

// NewAmqpPubSub - создание helper pub/sub
func NewAmqpPubSub(cfg amqp.ConnectionConfig) (*AmqpPubSub, error) {
	conn, err := amqp.NewConnection(cfg, watermill.NewStdLogger(true, true))

	if err != nil {
		return nil, err
	}

	return &AmqpPubSub{
		config: nil,
		conn:   conn,
	}, nil
}

// Publish - отправка сообщения
func (a *AmqpPubSub) Publish(payload interface{}, cfg ...Config) error {
	msg, err := messageForm(payload)

	if err != nil {
		return err
	}

	cc := &amqp.Config{}
	for _, opt := range cfg {
		opt(cc)
	}

	publisher, err := amqp.NewPublisherWithConnection(*cc, watermill.NewStdLogger(true, true), a.conn)

	if err != nil {
		return fmt.Errorf("error creating publisher: %w", err)
	}

	defer func() {
		if err := publisher.Close(); err != nil {
			fmt.Printf("error closing publisher: %v\n", err)
		}
	}()

	if err := publisher.Publish("", msg); err != nil {
		return fmt.Errorf("error publishing message: %w", err)
	}

	return nil
}

// RegisterHandler - Регистрирование консьюмеров для очередей
func (a *AmqpPubSub) RegisterHandler(handler structures.Handler, cfg ...Config) error {
	cc := &amqp.Config{}

	for _, opt := range cfg {
		opt(cc)
	}

	consumer, err := amqp.NewSubscriberWithConnection(*cc, watermill.NewStdLogger(true, true), a.conn)

	if err != nil {
		log.Printf("failed to subscribe to queue", err)

		return err
	}

	a.handlers = append(a.handlers, structures.ConsumeConfig{
		Consumer: consumer,
		Handler:  handler,
	})

	return nil
}

// Consume - запуск обработки консьюмеров
func (a *AmqpPubSub) Consume(ctx context.Context) error {
	errorChan := make(chan error, len(a.handlers))
	var wg sync.WaitGroup

	for _, consumeConfig := range a.handlers {
		wg.Add(1)

		go func(config structures.ConsumeConfig) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("panic error message: %v", r)
				}
			}()

			log.Printf("Try to subscribe with config: %+v", config)

			msgChannel, err := config.Consumer.Subscribe(ctx, "")

			if err != nil {
				errorChan <- fmt.Errorf("failed to consume subscribe: %w", err)
				return
			}

			log.Printf("Successfully subscribed")

			for {
				select {
				case <-ctx.Done():
					log.Println("Context cancelled, stopping consumer")
					return
				case msg, ok := <-msgChannel:
					if !ok {
						log.Println("Message channel closed")
						return
					}

					func(msg *message.Message) {
						defer func() {
							if r := recover(); r != nil {
								log.Printf("Panic error message: %v", r)
								msg.Ack()
							}
						}()

						handlerName := reflect.TypeOf(config.Handler).String()

						if reflect.TypeOf(config.Handler).Kind() == reflect.Func {
							pc := reflect.ValueOf(config.Handler).Pointer()
							fn := runtime.FuncForPC(pc)
							handlerName = fn.Name()
						}

						var parentCtx context.Context
						var span trace.Span

						if ctx != nil {
							if tracer.TraceClient != nil && tracer.TraceClient.IsEnabled {
								parentCtx, span = tracer.TraceClient.CreateSpan(ctx, "[Consumer handle]"+handlerName)
								defer span.End()
							}
						} else {
							parentCtx = ctx
						}

						err := config.Handler(parentCtx, msg)

						if err != nil {
							log.Printf("Error handling message: %v", err)
						}

						msg.Ack()
					}(msg)
				}
			}
		}(consumeConfig)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	var firstErr error
	for err := range errorChan {
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return firstErr
	}

	<-ctx.Done()

	return firstErr
}

// CloseConnection - закрытие коннекта с rabbitmq
func (a *AmqpPubSub) CloseConnection() error { return a.conn.Close() }

// MessageForm - Подготовка в формат сообщения для publish
func messageForm(notice interface{}) (*message.Message, error) {
	val, err := json.Marshal(notice)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal payload: %v", err)
	}
	msg := message.NewMessage(uuid.New().String(), val)

	return msg, nil
}
