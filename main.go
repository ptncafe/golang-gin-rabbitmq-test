package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ptncafe/golang-gin-rabbitmq-test/common/queue"
	"github.com/ptncafe/golang-gin-rabbitmq-test/controller"
	"github.com/ptncafe/golang-gin-rabbitmq-test/handler/queueHandler"
)

var queueClient queue.IQueueClient

func main() {
	queueClient = queue.NewQueueClient("amqp://services:services@172.30.119.91:5672/")
	subscribe := queueHandler.NewSubscribeQueue(queueClient)
	subscribe.RegisterSubscribeQueue()
	initGin()

}

func initGin() {
	r := gin.Default()
	router(r)
	r.Run() // listen and serve on 0.0.0.0:8080
}
func router(ginEngine *gin.Engine) {
	queueController := controller.NewQueueController(queueClient)

	ginEngine.GET("", controller.Ping)
	ginEngine.GET("/ping", controller.Ping)

	ginEngine.POST("/queue/publish", queueController.Publish)
}
