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
	docker logs rabbitmq

.PHONY: producer
producer:
	@go run producer/main.go

.PHONY: consumer
consumer:
	@go run consumer/main.go