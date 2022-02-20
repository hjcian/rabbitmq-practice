package main

import (
	"log"
	"rabbitmqpractice/rabbitmq"

	"github.com/google/uuid"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	handler, err := rabbitmq.New("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ/open a channel")
	defer handler.Close()

	exchangeName := "pub_sub_topic"

	body := uuid.NewString()
	log.Println("send a message:", body)

	failOnError(
		handler.Publish(exchangeName, []byte(body)),
		"Failed to publish a message",
	)

}
