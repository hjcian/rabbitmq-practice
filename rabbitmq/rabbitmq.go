package rabbitmq

import (
	"github.com/streadway/amqp"
)

type SubscriberOption func(s *subscriberOptions)

type subscriberOptions struct {
	routingKey string
}

func WithRoutingKey(routingKey string) SubscriberOption {
	return func(s *subscriberOptions) {
		s.routingKey = routingKey
	}
}

type ConsumerFunc func(message []byte, ack func())

type RabbitMQ interface {
	// New(url string) (*amqp.Channel, error)
	// Publish()

	// Subscribe creates a goroutine to subscribe message from channel.
	// Caller has responsibility to call ack() to acknowledge the message.
	Subscribe(exchangeName, subName string, fn ConsumerFunc, opts ...SubscriberOption) error
	Close()
}

func New(url string) (RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &rabbitMQ{conn, ch}, nil
}

type rabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (r *rabbitMQ) Close() {
	// order matters
	r.ch.Close()
	r.conn.Close()
}

func (r *rabbitMQ) Subscribe(exchangeName, subName string, consumerFn ConsumerFunc, optFns ...SubscriberOption) error {
	opts := subscriberOptions{
		routingKey: "#", // default to subscribe all topics. https://www.rabbitmq.com/tutorials/tutorial-five-go.html
	}
	for _, optFn := range optFns {
		optFn(&opts)
	}

	q, err := r.ch.QueueDeclare(
		subName, // name
		true,    // durable
		true,    // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	// Binding a Queue to a Exchange, see the philosophy explained here: https://www.rabbitmq.com/tutorials/tutorial-three-go.html
	if err := r.ch.QueueBind(
		q.Name,          // queue name
		opts.routingKey, // routing key
		exchangeName,    // exchange
		false,           // no-wait
		nil,             // arguments
	); err != nil {
		return err
	}

	msgs, err := r.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack set to false, need the consumer ack it explicitly
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			consumerFn(
				d.Body,
				func() { d.Ack(false) },
			)
		}
	}()

	return nil
}
