package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var (
	RabbitMQConn    *amqp091.Connection
	RabbitMQChannel *amqp091.Channel
)

const SeatCreationQueueName = "seat_creation_queue"
const BookingCancellationQueueName = "booking_cancellation_queue"

func InitRabbitMQ(amqpURL string) error {
	LogInfo("Attempting to connect to RabbitMQ at: %s", amqpURL)
	var err error
	var counts uint8 = 1
	const maxRetries = 10
	const retryDelay = 5 * time.Second

	for counts <= maxRetries {
		RabbitMQConn, err = amqp091.Dial(amqpURL)
		if err != nil {
			LogError("Attempt %d/%d: Failed to connect to RabbitMQ: %v. Retrying in %s...", counts, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
			counts++
			continue
		} else {
			LogInfo("Connected to RabbitMQ successfully!")
			break
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after %d retries: %v", maxRetries, err)
		return fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	LogInfo("Opening RabbitMQ channel...")
	RabbitMQChannel, err = RabbitMQConn.Channel()
	if err != nil {
		LogError("Failed to open RabbitMQ channel: %v", err)
		RabbitMQConn.Close()
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
		return fmt.Errorf("failed to open a RabbitMQ channel: %w", err)
	}

	if _, err := DeclareQueue(SeatCreationQueueName); err != nil {
		LogError("Failed to declare queue '%s': %v", SeatCreationQueueName, err)
		RabbitMQChannel.Close()
		RabbitMQConn.Close()
		log.Fatalf("Failed to declare queue '%s': %v", SeatCreationQueueName, err)
		return fmt.Errorf("failed to declare queue '%s': %w", SeatCreationQueueName, err)
	}

	if _, err := DeclareQueue(BookingCancellationQueueName); err != nil {
		LogError("Failed to declare queue '%s': %v", BookingCancellationQueueName, err)
		RabbitMQChannel.Close()
		RabbitMQConn.Close()
		log.Fatalf("Failed to declare queue '%s': %v", BookingCancellationQueueName, err)
		return fmt.Errorf("failed to declare queue '%s': %w", BookingCancellationQueueName, err)
	}

	LogInfo("RabbitMQ channel and all queues declared successfully!")
	return nil
}

func PublishMessage(exchange, routingKey string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := RabbitMQChannel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message to queue '%s': %w", routingKey, err)
	}
	LogInfo("Message published to exchange '%s' with routing key '%s'.", exchange, routingKey)
	return nil
}

func ConsumeMessages(queueName string) (<-chan amqp091.Delivery, error) {
	msgs, err := RabbitMQChannel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer for queue '%s': %w", queueName, err)
	}
	LogInfo("Started consuming messages from queue '%s'.", queueName)
	return msgs, nil
}

func DeclareQueue(queueName string) (amqp091.Queue, error) {
	q, err := RabbitMQChannel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return amqp091.Queue{}, fmt.Errorf("failed to declare queue '%s': %w", queueName, err)
	}
	LogInfo("Queue '%s' declared.", queueName)
	return q, nil
}

func SeatCreationQueue() string {
	return SeatCreationQueueName
}

func BookingCancellationQueue() string {
	return BookingCancellationQueueName
}

func CloseRabbitMQConnection() {
	if RabbitMQChannel != nil {
		RabbitMQChannel.Close()
		LogInfo("RabbitMQ channel closed.")
	}
	if RabbitMQConn != nil {
		RabbitMQConn.Close()
		LogInfo("RabbitMQ connection closed.")
	}
}
