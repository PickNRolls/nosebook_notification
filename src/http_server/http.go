package http_server

import (
	"log"
	"net/http"
	"notification/src/rabbitmq"
	"notification/src/socket"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	original *http.Server
	hub      *socket.Hub
	rmq      *rabbitmq.Handlers
}

func New(hub *socket.Hub, rmq *rabbitmq.Handlers) *Server {
	out := &Server{
		hub: hub,
		rmq: rmq,
		original: &http.Server{
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	out.original.Handler = http.HandlerFunc(out.Handler)

	return out
}

func (this *Server) Run(addr string) {
	this.original.Addr = addr
	log.Println("Server is listening on " + this.original.Addr)
	this.original.ListenAndServe()
}

func (this *Server) Handler(w http.ResponseWriter, req *http.Request) {
	client := socket.NewClient(this.hub, this.rmq)

	authUserIdHeaders := req.Header["X-Auth-User-Id"]
	userId, err := uuid.Parse(authUserIdHeaders[0])
	if err != nil {
		log.Fatalln(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	client.Run(w, req, userId)
}
