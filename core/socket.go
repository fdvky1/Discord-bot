package core

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
)

var WSClients WebSocketClients

type WebSocketClients struct {
	Clients map[string]*websocket.Conn
}

type Msg struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func SendLog(id string, log string) {
	fmt.Printf("%s <-- %s\n", id, log)
	client := WSClients.Clients[id]
	if client != nil {
		client.WriteJSON(&Msg{
			Type:  "log",
			Value: log,
		})
	}
}
