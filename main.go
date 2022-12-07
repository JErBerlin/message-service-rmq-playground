package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const (
	// Replace with your own RabbitMQ connection URL
	rabbitMQURL = "amqp://guest:guest@localhost:5672"

	// Replace with your own chat queue name
	chatQueue = "chat"
)

// User represents a user in the chat service
type User struct {
	username string
	password string
}

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// Declare the chat queue
	q, err := ch.QueueDeclare(
		chatQueue, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a map of registered users
	users := make(map[string]User)

	// Register a new user with a username and password
	registerUser := func(username, password string) error {
		if _, ok := users[username]; ok {
			return fmt.Errorf("username %q is already taken", username)
		}
		users[username] = User{username, password}
		return nil
	}

	// Authenticate a user with a username and password
	authenticateUser := func(username, password string) error {
		user, ok := users[username]
		if !ok {
			return fmt.Errorf("username %q not found", username)
		}
		if user.password != password {
			return fmt.Errorf("incorrect password for username %q", username)
		}
		return nil
	}

	// Consume messages from the chat queue
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatal(err)
	}

	// Handle incoming messages
	go func() {
		for msg := range msgs {
			log.Printf("Received message: %s", msg.Body)
		}
	}()

	// Publish a message to the chat queue
	publishMessage := func(username, password, text string) error {
		if err := authenticateUser(username, password); err != nil {
			return err
		}
		err := ch.Publish(
			"",        // exchange
			chatQueue, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(text),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
		log.Printf("Sent message: %s", text)
		return nil
	}

	// Register a new user
	if err := registerUser("alice", "password123"); err != nil {
		log.Fatal(err)
	}

	// Publish a message as the registered user
	if err := publishMessage("alice", "password123", "Hello, world!"); err != nil {
		log.Fatal(err)
	}
}
