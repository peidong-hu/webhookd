package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/vision-it/webhookd/config"
	. "github.com/vision-it/webhookd/logging"
)

var mqconfig config.MQConfig
var ch *amqp.Channel

func Connect(c config.MQConfig) (*amqp.Connection, *amqp.Channel) {
	mqconfig = c
	mqConnection := fmt.Sprintf("%s://%s:%s@%s:%d/",
		c.Protocol, c.User, c.Password, c.Host, c.Port)

	conn, err := amqp.Dial(mqConnection)
	FailOnError(err, "Failed to connect to RabbitMQ")

	ch, err = conn.Channel()
	FailOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		c.Exchange, // name
		"fanout",   // type
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	FailOnError(err, "Failed to declare an exchange")

	return conn, ch
}

func publishMessage(ch *amqp.Channel, message string) (err error) {

	err = ch.Publish(
		mqconfig.Exchange, // exchange
		"",                // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(message),
		},
	)
	if err != nil {
		Lg(0, "Failed to declare exchange %s: %s", mqconfig.Exchange, err)
	}

	Lg(2, "Published message %s to exchange %s", message, mqconfig.Exchange)

	return err
}

func Publish(message string, exchange string) (err error) {
	err = ch.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(message),
		},
	)

	return err
}
