package controller

import (
	"encoding/json"

	"github.com/ptncafe/golang-gin-rabbitmq-test/common/constants"

	"github.com/gin-gonic/gin"
	"github.com/ptncafe/golang-gin-rabbitmq-test/common/queue"
)

type queueController struct {
	queueClient queue.IQueueClient
}

func NewQueueController(queueClient queue.IQueueClient) *queueController {
	return &queueController{queueClient: queueClient}
}
func (qc *queueController) Publish(c *gin.Context) {
	var data map[string]interface{}
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	byteData, err := json.Marshal(data)
	qc.queueClient.PublishOnQueue(constants.DefaultQueueName, byteData)

	c.JSON(200, gin.H{
		"message": "",
		"data":    data,
	})
	return
}
