package main

import (
	"log"
	"notification/src/http_server"
	"notification/src/rabbitmq"
	"notification/src/socket"

	"github.com/google/uuid"
)

func main() {
	rmq := rabbitmq.Connect()
	defer rmq.Close()

	hub := socket.NewHub(rmq)

	server := http_server.New(hub, rmq)
	go server.Run("0.0.0.0:8081")

	msgs, err := rmq.Channel.Consume(rmq.Queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalln("Failed to register a consumer")
		return
	}

	for d := range msgs {
		userId, err := uuid.Parse(d.RoutingKey)
		if err != nil {
			log.Fatalln("Failed to parse routing key as user uuid")
			continue
		}

		log.Printf("Received a message: %s, user_id: %v", d.Body, userId)

		clients := hub.UserClients(userId)
		for _, client := range clients {
			client.Send() <- d.Body
		}

		if len(clients) > 0 {
			log.Printf("Message was sent to all user(id:%v) clients\n", userId)
			log.Printf("Total amount of clients: %v\n", len(clients))
		} else {
			log.Printf("Got message, but no client for this user(id:%v)\n", userId)
		}
	}
}
