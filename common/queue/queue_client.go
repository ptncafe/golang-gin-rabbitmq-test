package queue

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// Defines our interface for connecting, producing and consuming messages.
type IQueueClient interface {
	ConnectToBroker(connectionString string)
	//Publish(msg []byte, exchangeName string, exchangeType string) error
	PublishOnQueue(queueName string, body []byte) error
	// Subscribe(exchangeName string, exchangeType string, consumerName string, handlerFunc func(amqp.Delivery)) error
	SubscribeToQueue(queueName string, consumerName string, handlerFunc func(amqp.Delivery)) error
	Close()
}

func NewQueueClient(connectionString string) *queueClient {
	QueueClient := queueClient{}
	QueueClient.ConnectToBroker(connectionString)
	return &QueueClient
}

// Real implementation, encapsulates a pointer to an amqp.Connection
type queueClient struct {
	conn *amqp.Connection
}

func (m *queueClient) ConnectToBroker(connectionString string) {
	if connectionString == "" {
		panic("Cannot initialize connection to broker, connectionString not set. Have you initialized?")
	}

	var err error
	m.conn, err = amqp.Dial(fmt.Sprintf("%s/", connectionString))
	if err != nil {
		panic("Failed to connect to AMQP compatible broker at: " + connectionString)
	}
}

func (m *queueClient) PublishOnQueue(queueName string, body []byte) error {
	if m.conn == nil {
		panic("Tried to send message before connection was initialized. Don't do that.")
	}
	ch, err := m.conn.Channel() // Get a channel from the connection
	defer ch.Close()

	// Declare a queue that will be created if not exists with some args
	queue, err := ch.QueueDeclare(
		queueName, // our queue name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	// Publishes a message onto the queue.
	err = ch.Publish(
		"",         // use the default exchange
		queue.Name, // routing key, e.g. our queue name
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body, // Our JSON body as []byte
		})
	fmt.Printf("A message was sent to queue %v: %v", queueName, body)
	return err
}

func (m *queueClient) SubscribeToQueue(queueName string, consumerName string, handlerFunc func(amqp.Delivery)) error {
	ch, err := m.conn.Channel()
	failOnError(err, "Failed to open a channel")

	log.Printf("Declaring Queue (%s)", queueName)
	queue, err := ch.QueueDeclare(
		queueName, // name of the queue
		false,     // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	failOnError(err, "Failed to register an Queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		queue.Name,   // queue
		consumerName, // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")
	go consumeLoop(msgs, handlerFunc)
	return nil
}

func (m *queueClient) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}

func consumeLoop(deliveries <-chan amqp.Delivery, handlerFunc func(d amqp.Delivery)) {
	for d := range deliveries {
		// Invoke the handlerFunc func we passed as parameter.
		handlerFunc(d)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
