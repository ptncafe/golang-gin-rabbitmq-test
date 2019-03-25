package queueHandler

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

func subscribeDefaultQueue(delivery amqp.Delivery) {
	fmt.Printf("START SubscribeDefaultQueue: %v\n %v\n", string(delivery.Body), time.Now().Format("2006-01-02 15:04:05"))

	time.Sleep(10000 * time.Millisecond)
	fmt.Printf("END SubscribeDefaultQueue: %v\n", string(delivery.Body))
}
