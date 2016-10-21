package wskeleton

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
)

type client struct {
	// userID string
	ws                *websocket.Conn
	send              chan Message
	processRawMessage func([]byte) Message
}

func NewClient(con *websocket.Conn, fn func([]byte) Message) client {
	return client{
		// userID: user,
		ws:                con,
		send:              make(chan Message, maxMessageSize),
		processRawMessage: fn,
	}
}

func (c *client) ReadMe(h *Hub) {
	defer func() {
		h.Unregister(c)
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		h.broadcast <- c.processRawMessage(message)
	}
}

func (c *client) SendMe(h *Hub) {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		h.unregister <- c
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, message)
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *client) write(mt int, msg Message) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteJSON(msg)
}
