package main

import (
	"log"
	"net/http"
	"strings"

	"excinity/exchange"
	"excinity/routes"
	"excinity/services"
)

var (
	exchangeClient exchange.ExchangeClient
)

func main() {

	exchangeClient = exchange.NewBinanceClient()
	symbols, err := exchangeClient.GetAvailableSymbols()
	if err != nil {
		log.Fatal(err)
	}

	for _, symbol := range symbols {
		go services.StartSymbolStream(exchangeClient, strings.ToLower(symbol))
	}

	// for _, symbol := range symbols {
	// 	go startBinanceClient(strings.ToLower(symbol))
	// }

	http.HandleFunc("/ws", routes.HandleWebsocketConnections)

	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
