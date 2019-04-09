package queue

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const exchangeDefault = "golang-gin-rabbitmq-test-exchangeDefault"

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
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
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
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body, // Our JSON body as []byte
		})
	fmt.Printf("A message was sent to queue %v: %v", queueName, body)
	return err
}

func (m *queueClient) SubscribeToQueue(queueName string, consumerName string, handlerFunc func(amqp.Delivery)) error {
	var err error
	log.Printf("got Connection, getting Channel")
	m.channel, err = m.conn.Channel()
	if err != nil {
		return err
	}
	defer m.channel.Close()
	log.Printf("got Channel, declaring Exchange (%s)", "direct")
	if err = m.channel.ExchangeDeclare(
		exchangeDefault, // name of the exchange
		"direct",        // type flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
		false,           // durable
		false,           // delete when complete
		false,           // internal
		false,           // noWait
		nil,             // arguments
	); err != nil {
		return err
	}

	log.Printf("declared Exchange, declaring Queue (%s)", queueName)
	state, err := m.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return err
	}
	var key = queueName + exchangeDefault
	log.Printf("declared Queue (%d messages, %d consumers), binding to Exchange (key '%s')",
		state.Messages, state.Consumers, key)

	if err = m.channel.QueueBind(
		queueName,       // name of the queue
		key,             // bindingKey
		exchangeDefault, // sourceExchange
		false,           // noWait
		nil,             // arguments
	); err != nil {
		return err
	}
	m.channel.Qos(10, 0, false)
	log.Printf("Queue bound to Exchange, starting Consume (consumer tag '%s')", "tagConsume")
	deliveries, err := m.channel.Consume(
		queueName,    // name
		"tagConsume", // consumerTag,
		false,        //auto-ack or noack
		false,        // exclusive
		false,        // noLocal
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return err
	}

	done := make(chan bool)

	go func() {
		msgCount := 0
		for d := range deliveries {
			fmt.Printf("\nMessage Count: %d, Message Body: %s\n", msgCount, d.Body)
			msgCount++
			handlerFunc(d)
			d.Ack(false)
		}
		log.Printf("handle: deliveries channel closed")
	}()
	<-done
	return nil
}

func (m *queueClient) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
