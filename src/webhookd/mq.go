package main

import (
	"github.com/streadway/amqp"
	"fmt"
)

const MQMessageVersion string = "0.0"

var MQCONFIG MQConfig

type MQMessage struct {
	Version string `json:"version"`
	Repository string `json:"repository"`
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	Message string `json:"message"`
	Author string `json:"author"`
	Trigger string `json:"trigger"`
}

func connectMQ(c MQConfig) (conn *amqp.Connection, ch *amqp.Channel){
	MQCONFIG = c
	mqConnection := fmt.Sprintf("%s://%s:%s@%s:%d/",
		c.Protocol, c.User, c.Password, c.Host, c.Port)

	conn, err := amqp.Dial(mqConnection)
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		c.Exchange, // name
		"fanout", // type
		false, // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	return conn, ch
}


func publishMessage(ch *amqp.Channel, message string) (err error){

	err = ch.Publish(
		MQCONFIG.Exchange, // exchange
		"", // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(message),
		},
	)
	if err != nil {
		lg(0, "Failed to declare exchange %s: %s", MQCONFIG.Exchange, err)
	}

	lg(2, "Published message %s to exchange %s", message, MQCONFIG.Exchange)

	return err
}
