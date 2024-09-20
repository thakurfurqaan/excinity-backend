package main

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

var symbols = []string{"BTCUSDT", "ETHUSDT", "PEPEUSDT"}

func main() {
	// Start Binance WebSocket clients
	for _, symbol := range symbols {
		go startBinanceClient(symbol)
	}

	// HTTP handler for WebSocket connections
	http.HandleFunc("/ws", handleConnections)

	// Serve static files
	fs := http.FileServer(http.Dir("./frontend/build"))
	http.Handle("/", fs)

	log.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
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

func broadcastCandle(candle Candle) {
	clientsMux.Lock()
	for client := range clients {
		err := client.WriteJSON(candle)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
	clientsMux.Unlock()
}
