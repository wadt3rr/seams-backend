package ws

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {

	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {

		var msg Message

		err := c.conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// ответ бота
		reply := GetBotReply(msg.Text)

		response := ServerMessage{
			Type: "bot",
			Text: reply,
		}

		data, err := json.Marshal(response)
		if err != nil {
			continue
		}

		c.send <- data
	}
}

func (c *Client) WritePump() {
	defer c.conn.Close()

	for {
		select {

		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (c *Client) HubRegister() {
	c.hub.register <- c
}
