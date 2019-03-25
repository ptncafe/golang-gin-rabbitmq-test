package queueHandler

import (
	"github.com/ptncafe/golang-gin-rabbitmq-test/common/constants"
	"github.com/ptncafe/golang-gin-rabbitmq-test/common/queue"
)

type subscribeQueue struct {
	queueClient queue.IQueueClient
}

func NewSubscribeQueue(queueClient queue.IQueueClient) *subscribeQueue {
	return &subscribeQueue{queueClient: queueClient}
}

func (sq *subscribeQueue) RegisterSubscribeQueue() {
	sq.queueClient.SubscribeToQueue(constants.DefaultQueueName, "test-services", subscribeDefaultQueue)
}
