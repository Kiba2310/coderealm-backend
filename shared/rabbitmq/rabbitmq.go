package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// Connect establishes a connection to RabbitMQ with retry logic
func Connect(url string) (*Client, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v. Retrying in 5 seconds...", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("could not open channel: %w", err)
	}

	// Ensure the common exchanges exist
	err = ch.ExchangeDeclare(
		"events", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("could not declare exchange: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// Publish sends an event to the 'events' exchange
func (c *Client) Publish(routingKey string, body interface{}) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	err = c.channel.Publish(
		"events",   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         bytes,
			DeliveryMode: amqp.Persistent,
		},
	)
	return err
}

// Subscribe listens to a specific routing key and calls the handler
func (c *Client) Subscribe(queueName, routingKey string, handler func([]byte) error) error {
	// Declare the queue
	q, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("could not declare queue: %w", err)
	}

	// Bind the queue to the exchange
	err = c.channel.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		"events",   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("could not bind queue: %w", err)
	}

	// Start consuming
	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("could not register consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			err := handler(d.Body)
			if err != nil {
				log.Printf("Error handling message: %v", err)
				// Nack and requeue
				d.Nack(false, true)
			} else {
				// Ack
				d.Ack(false)
			}
		}
	}()

	return nil
}
