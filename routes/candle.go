package routes

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
)

func HandleWebsocketConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clientsMux.Lock()
	clients[ws] = true
	clientsMux.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			clientsMux.Lock()
			delete(clients, ws)
			clientsMux.Unlock()
			break
		}
	}
}

func BroadcastData(data interface{}) {
	clientsMux.Lock()
	for client := range clients {
		err := client.WriteJSON(data)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
