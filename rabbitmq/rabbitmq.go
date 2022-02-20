package rabbitmq

import (
	"github.com/streadway/amqp"
)

//
// Options for subscriber
//

type SubscriberOption func(s *subscriberOptions)

type subscriberOptions struct {
	// NOTE:
	// for avoid the confusion with Channel.Publish parameter,
	// the official tutorial calls the relationship between queue and exchange to "binding key".
	// ref here: https://www.rabbitmq.com/tutorials/tutorial-four-go.html
	bindingKey string
}

func WithBindingKey(bindingKey string) SubscriberOption {
	return func(s *subscriberOptions) {
		s.bindingKey = bindingKey
	}
}

//
// Options for publisher
//

type PublisherOption func(p *publisherOptions)

type publisherOptions struct {
	contentType string
	routingKey  string
}

func WithContentType(ct string) PublisherOption {
	return func(p *publisherOptions) {
		p.contentType = ct
	}
}

func WithRoutingKey(key string) PublisherOption {
	return func(p *publisherOptions) {
		p.routingKey = key
	}
}

//
// Common definitions
//

type ConsumerFunc func(message []byte, ack func())

//
// Interface and concret struct implementation
//

type RabbitMQ interface {
	// Publish sends the message to broker.
	Publish(exchangeName string, payload []byte, optFns ...PublisherOption) error

	// Subscribe creates a goroutine to subscribe message from channel.
	// Caller has responsibility to call ack() to acknowledge the message.
	Subscribe(exchangeName, subName string, fn ConsumerFunc, opts ...SubscriberOption) error

	// Close closes all connected resources, and result in the internal channel being closed.
	// An application server should implement graceful shutdown by calling Close() before exiting.
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
	// closing order is matters. first open, last close
	r.ch.Close()
	r.conn.Close()
}

func (r *rabbitMQ) Publish(exchangeName string, payload []byte, optFns ...PublisherOption) error {
	opts := publisherOptions{
		contentType: "",
		routingKey:  "", // default is empty string, the counterpart of Subscribe default binding key is "#"
	}
	for _, optFn := range optFns {
		optFn(&opts)
	}

	if err := r.ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		return nil
	}

	return r.ch.Publish(
		exchangeName,    // exchange
		opts.routingKey, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: opts.contentType,
			Body:        []byte(payload),
		})
}

func (r *rabbitMQ) Subscribe(exchangeName, subName string, consumerFn ConsumerFunc, optFns ...SubscriberOption) error {
	opts := subscriberOptions{
		bindingKey: "#", // default to subscribe all topics. https://www.rabbitmq.com/tutorials/tutorial-five-go.html
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
		opts.bindingKey, // routing key for binding queue and exchange
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
