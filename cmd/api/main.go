package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"excinity/exchange"
	"excinity/routes"
	"excinity/services"
)

func main() {
	exchangeClient := exchange.NewBinanceClient()
	symbols, err := exchangeClient.GetAvailableSymbols()
	if err != nil {
		log.Fatal(err)
	}

	for _, symbol := range symbols {
		go func(s string) {
			err := services.StartSymbolStream(exchangeClient, strings.ToLower(s))
			if err != nil {
				log.Printf("Error starting stream for symbol %s: %v", s, err)
			}
		}(symbol)
	}

	http.HandleFunc("/ws", routes.HandleWebsocketConnections)

	go func() {
		log.Println("Server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Keep the main goroutine running
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
}
