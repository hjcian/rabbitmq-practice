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

.PHONY: producer
producer:
	@go run producer/main.go

.PHONY: consumer
consumer:
	@go run consumer/main.go

.PHONY: pubsub_producer, pubsub_consumer
pubsub_producer:
	@go run pubsub_example/producer/producer.go

pubsub_consumer:
	@go run pubsub_example/consumer/consumer.go

.PHONY: list_q_x_b
list_q_x_b:
	@docker exec -it rabbitmq rabbitmqctl list_queues
	@docker exec -it rabbitmq rabbitmqctl list_exchanges
	@docker exec -it rabbitmq rabbitmqctl list_bindings