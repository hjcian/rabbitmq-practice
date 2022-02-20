package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func workerSet(wg *sync.WaitGroup, setNum int) {
	wg.Add(1)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		fmt.Sprintf("%v-workerset", setNum), // name
		false,                               // durable
		false,                               // delete when unused
		false,                               // exclusive
		false,                               // no-wait
		nil,                                 // arguments
	)
	log.Printf("queue name created: %s", q.Name)
	ch.QueueBind(
		q.Name,        // queue name
		"",            // routing key
		"pub_sub_hub", // exchange
		false,
		nil,
	)

	failOnError(err, "Failed to declare a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		log.Println(setNum, "start worker 1")
		for d := range msgs {
			log.Printf("[%v-1] Received a message: %s", setNum, d.Body)
		}
		log.Println(setNum, "worker 1 exit")
	}()

	go func() {
		log.Println(setNum, "start worker 2")
		for d := range msgs {
			log.Printf("[%v-2] Received a message: %s", setNum, d.Body)
		}
		log.Println(setNum, "worker 2 exit")
	}()
	wg.Wait()
	log.Println(setNum, "workerSet exit")
}

func main() {

	exitSig := make(chan os.Signal, 1)
	signal.Notify(exitSig, os.Interrupt)

	var wg sync.WaitGroup
	go workerSet(&wg, 1)
	go workerSet(&wg, 2)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-exitSig
	wg.Done()
	wg.Done()
}
