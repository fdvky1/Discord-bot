package core

import (
	"fmt"
	"strings"

	"github.com/gofiber/contrib/websocket"
)

var WSClients WebSocketClients
var Cached CachedType

type WebSocketClients struct {
	Clients map[string]*websocket.Conn
}

type CachedType map[string]string

type Msg struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func FindIdByToken(token string) string {
	// idk why the query isnt working, I think this issue is related to https://github.com/nedpals/supabase-go/issues/28
	// var response []struct {
	// 	Id string `json:"id"`
	// }
	// err := Supabase.DB.From("users").Select("*").Eq("bot_token", strings.Split(token, " ")[1]).Execute(&response)
	// if err != nil {
	// 	panic(err)
	// }
	// return response[0].Id
	return Cached[strings.Split(token, " ")[1]]
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
