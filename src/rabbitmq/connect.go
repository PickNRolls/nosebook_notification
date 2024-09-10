package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError[T any](data T, err error) T {
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	return data
}

func Connect() *Handlers {
	conn := failOnError(amqp.Dial("amqp://guest:guest@rabbitmq:5672/"))
	ch := failOnError(conn.Channel())

	err := ch.ExchangeDeclare("notifications", "direct", false, false, false, false, nil)
	failOnError(struct{}{}, err)
	queue := failOnError(ch.QueueDeclare("", false, false, false, false, nil))

	return &Handlers{
		Conn:    conn,
		Channel: ch,
		Queue:   queue,
	}
}
