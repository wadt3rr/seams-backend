package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"seams-backend/internal/ws"
)

type SupportWS struct {
	hub *ws.Hub
}

func NewSupportWS(hub *ws.Hub) *SupportWS {
	return &SupportWS{
		hub: hub,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *SupportWS) Handle(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade error:", err)
		return
	}

	client := ws.NewClient(h.hub, conn)

	// регистрация клиента
	client.HubRegister()

	go client.WritePump()
	go client.ReadPump()
}
