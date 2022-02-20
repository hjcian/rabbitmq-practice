package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"rabbitmqpractice/rabbitmq"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func workerSet(handler rabbitmq.RabbitMQ, setNum int) {
	handler.Subscribe(
		"pub_sub_hub",
		fmt.Sprintf("workerSet-%v", setNum),
		func(msg []byte, ack func()) {
			log.Printf("[%v-1] Received a message: %s ...", setNum, msg)
			time.Sleep(time.Second * 5)
			log.Printf("[%v-1] Finished message: %s", setNum, msg)
			ack()
		},
	)
	handler.Subscribe(
		"pub_sub_hub",
		fmt.Sprintf("workerSet-%v", setNum),
		func(msg []byte, ack func()) {
			log.Printf("[%v-2] Received a message: %s ...", setNum, msg)
			time.Sleep(time.Second * 5)
			log.Printf("[%v-2] Finished message: %s", setNum, msg)
			ack()
		},
	)
}

func main() {
	exitSig := make(chan os.Signal, 1)
	signal.Notify(exitSig, os.Interrupt)

	handler, err := rabbitmq.New("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ/open a channel")
	workerSet(handler, 999)
	// for i := range make([]struct{}, 1) {
	// 	go workerSet(handler, i)
	// }

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
	<-exitSig
	handler.Close()
	log.Printf("[*] exit demo, bye.")
}
