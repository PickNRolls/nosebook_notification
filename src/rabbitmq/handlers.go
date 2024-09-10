package rabbitmq

import "github.com/rabbitmq/amqp091-go"

type Handlers struct {
	Conn    *amqp091.Connection
	Queue   amqp091.Queue
	Channel *amqp091.Channel
}

func (this *Handlers) Close() {
	this.Conn.Close()
	this.Channel.Close()
}
