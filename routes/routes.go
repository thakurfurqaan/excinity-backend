package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"excinity/config"
	"excinity/services"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SetupRoutes(aggregationService *services.AggregationService) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/ws/{symbol}", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, aggregationService)
	})

	r.HandleFunc("/api/history/{symbol}", func(w http.ResponseWriter, r *http.Request) {
		getHistoricalData(w, r, aggregationService)
	})

	return r
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, aggregationService *services.AggregationService) {
	symbol := mux.Vars(r)["symbol"]

	conn, err := upgradeConnection(w, r)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	streamCandles(conn, symbol, aggregationService)
}

func upgradeConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return upgrader.Upgrade(w, r, nil)
}

func streamCandles(conn *websocket.Conn, symbol string, aggregationService *services.AggregationService) {
	for {
		candle, err := aggregationService.GetLatestCandle(symbol)
		if err != nil {
			sendError(conn, err)
			return
		}

		if err := sendCandle(conn, candle); err != nil {
			return
		}

		time.Sleep(config.STREAM_UPDATE_INTERVAL)
	}
}

func sendError(conn *websocket.Conn, err error) {
	conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
}

func sendCandle(conn *websocket.Conn, candle interface{}) error {
	return conn.WriteJSON(candle)
}

func getHistoricalData(w http.ResponseWriter, r *http.Request, aggregationService *services.AggregationService) {
	symbol := mux.Vars(r)["symbol"]
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	candles, err := aggregationService.GetHistoricalData(symbol, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(candles)
}
