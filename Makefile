MQ_PORT?=5672

stop-rabbitmq:
	@echo "[`date`] Stopping previous launched RabbitMQ [if any]"
	docker stop rabbitmq || true

rabbitmq:
	@echo "[`date`] Starting RabbitMQ container"
	docker run -d --rm --name rabbitmq\
		-p ${MQ_PORT}:5672 \
		rabbitmq:3.8

restart-rabbitmq: stop-rabbitmq
restart-rabbitmq: rabbitmq

log-rabbitmq:
	docker logs -f rabbitmq

.PHONY: helloworld_producer, helloworld_consumer
helloworld_producer:
	@go run hello_word/producer/producer.go

helloworld_consumer:
	@go run hello_word/consumer/consumer.go

.PHONY: pubsub_producer, pubsub_consumer
pubsub_producer:
	@go run pubsub/producer/producer.go

pubsub_consumer:
	@go run pubsub/consumer/consumer.go

.PHONY: list_q_x_b
list_q_x_b:
	@docker exec -it rabbitmq rabbitmqctl list_queues
	@docker exec -it rabbitmq rabbitmqctl list_exchanges
	@docker exec -it rabbitmq rabbitmqctl list_bindings