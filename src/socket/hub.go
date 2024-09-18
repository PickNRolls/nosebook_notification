package socket

import (
	"log"
	"notification/src/rabbitmq"

	"github.com/google/uuid"
)

type Hub struct {
	clients map[uuid.UUID][]*Client
	rmq     *rabbitmq.Handlers
}

func NewHub(rmq *rabbitmq.Handlers) *Hub {
	return &Hub{
		clients: map[uuid.UUID][]*Client{},
		rmq:     rmq,
	}
}

func (this *Hub) UserClients(userId uuid.UUID) []*Client {
	return this.clients[userId]
}

func (this *Hub) Subscribe(client *Client) {
	err := this.rmq.Channel.QueueBind(this.rmq.Queue.Name, client.userId.String(), "notifications", false, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}

	if _, has := this.clients[client.userId]; !has {
		this.clients[client.userId] = []*Client{}
	}
	this.clients[client.userId] = append(this.clients[client.userId], client)

	log.Printf("Subscribe new client for user(id:%v)\n", client.userId)
	log.Printf("Total amount of clients for user(id:%v) = %v\n", client.userId, len(this.clients[client.userId]))
}

func (this *Hub) Unsubscribe(userId uuid.UUID, client *Client) {
	clients := this.clients[userId]
	if clients == nil {
		return
	}

	index := -1
	for i, c := range clients {
		if c == client {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}

	log.Printf("Unsubscribe client for user(id:%v)\n", userId)

	clients[index] = clients[len(clients)-1]
	this.clients[userId] = clients[:len(clients)-1]

	if len(this.clients[userId]) == 0 {
		delete(this.clients, userId)

		log.Printf("Removing binding=%v between exchange \"notifications\", no more clients for it\n", client.userId.String())
		err := this.rmq.Channel.QueueUnbind(this.rmq.Queue.Name, client.userId.String(), "notifications", nil)
		if err != nil {
			log.Fatalln(err)
		}
	}
	close(client.Send())

	log.Printf("Total amount of clients for user(id:%v) = %v\n", client.userId, len(this.clients[client.userId]))
}

type BroadcastFilter struct {
	UserId uuid.UUID
}

func (this *Hub) Broadcast(message []byte, filter *BroadcastFilter) {
	for userId, userClients := range this.clients {
		if filter != nil && userId != filter.UserId {
			continue
		}

		for _, client := range userClients {
			select {
			case client.Send() <- message:

			default:
				this.Unsubscribe(userId, client)
			}
		}
	}
}
