package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	symbols = []string{"AAPL", "AMZN"}
	// symbols := []string{"AAPL", "AMZN", "BINANCE:BTCUSDT", "IC MARKETS:1"}

	// Broadcast messages to all connected clients
	broadcast = make(chan *BroadCastMessage)

	//Map all ongoing clients and symbols they are subscribed to
	clientConns = make(map[*websocket.Conn]string)

	//map of all ongoing live candles for each symbol
	tempCandles = make(map[string]*TempCandle)

	mu sync.Mutex
)

func main() {
	//Env config
	env := EnvConfig()

	// db connection
	db := DBConnection(env)

	// connect to finnhub websocket
	finnhubWSConn := connectToFinnhub(env)
	defer finnhubWSConn.Close()

	// handle finnhub incomming message
	go handleFinnhubMessages(finnhubWSConn, db)

	// Broadcast candle update to all clients connected
	go broadcastUpdates()

	//--- Endpoints---
	//Connect to the websocket
	http.HandleFunc("/ws", wsHandler)

	//fetch all past candle for all of the symbols
	http.HandleFunc("/stocks-history", func(w http.ResponseWriter, r *http.Request) {
		StockHistoryHandler(w, r, db)
	})

	//fetch all past candle for specific symbol
	http.HandleFunc("/stocks-candles", func(w http.ResponseWriter, r *http.Request) {
		CandleHandler(w, r, db)
	})

	//Serve the endpoints
	http.ListenAndServe(fmt.Sprintf(":%s", env.SERVER_PORT), nil)
}

// websocket endpoints to connect to the latest updates on the symbols they're subscribe to
func wsHandler(w http.ResponseWriter, r *http.Request) {
	//Update incomming GET Request into a websocket connection
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade conncection: %s", err)
	}

	// close ws connection & unregister client when they disconnect
	defer conn.Close()
	defer func() {
		delete(clientConns, conn)
		log.Println("Client disconnected")
	}()

	//Register the new client to the symbols they are subscribe to
	for {
		_, symbols, err := conn.ReadMessage()
		clientConns[conn] = string(symbols)
		log.Println("New client connected")

		if err != nil {
			log.Println("Error reading from the client: ", err)
			break
		}
	}
}

// fetch all past candle for all of the symbols
func StockHistoryHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	//Query the db for all candle data from all symbols
	var candles []Candle
	db.Order("timestamp asc").Find(&candles)

	// create a map to group data by symbols
	groupedData := make(map[string][]Candle)

	// Group the candles by symbols
	for _, candle := range candles {
		symbol := candle.Symbol
		groupedData[symbol] = append(groupedData[symbol], candle)
	}

	//Marshal the groupedData to Json and send over http
	jsonResponse, _ := json.Marshal(groupedData)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// fetch all past candle for specific symbol
func CandleHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	//Get the symbol from the Query
	symbol := r.URL.Query().Get("symbol")

	//Query the db for all candle data for that symbol
	var candles []Candle
	db.Where("symbol = ?", symbol).Order("timestamp asc").Find(&candles)

	//Marshal the candle data into JSON and send over http
	jsonResponse, _ := json.Marshal(candles)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// connect to finnhub websocket
func connectToFinnhub(env *Env) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://ws.finnhub.io?token=%s", env.API_KEY), nil)
	if err != nil {
		panic(err)
	}

	for _, s := range symbols {
		msg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": s})
		ws.WriteMessage(websocket.TextMessage, msg)
	}

	return ws
}

// handle finnhub incomming message
func handleFinnhubMessages(ws *websocket.Conn, db *gorm.DB) {
	finnhubMessage := &FinnhubMessage{}

	for {
		if err := ws.ReadJSON(finnhubMessage); err != nil {
			fmt.Println("Error reading the message: ", err)
			continue
		}

		// only try to process the message data if it's a trade operation
		if finnhubMessage.Type == "trade" {
			for _, trade := range finnhubMessage.Data {
				//process the trade data
				processTradeData(&trade, db)
			}
		}

		//Clean up old trade older than 20 min
		cutoffTime := time.Now().Add(-20 * time.Minute)
		db.Where("timestamp < ?", cutoffTime).Delete(&Candle{})
	}
}

// Process each trade and update or create temporary candle
func processTradeData(trade *TradeData, db *gorm.DB) {
	//Protect the go routine from data races
	mu.Lock()
	defer mu.Unlock()

	//Extract trade data
	symbol := trade.Symbol
	price := trade.Price
	volume := float64(trade.Volume)
	timestamp := time.UnixMilli(trade.Timestamp)

	// Retrive or create a tempCandles for the symbols
	tempCandle, exists := tempCandles[symbol]

	//if the tempCandle does not exists or should be already closed
	if !exists || timestamp.After(tempCandle.CloseTime) {
		//Finalize and save the previous candle, start a new one
		if exists {
			// convert the tempCandles to a candle
			candle := tempCandle.toCandle()

			//Save the candle to a db
			if err := db.Create(candle).Error; err != nil {
				fmt.Println("Error saving the candle to the db: ", err)
			}

			//Broadcast the close candle
			broadcast <- &BroadCastMessage{
				UpdateType: Closes,
				Candle:     candle,
			}
		}

		//Initialize a new candle
		tempCandle = &TempCandle{
			Symbol:     symbol,
			OpenTime:   timestamp,
			CloseTime:  timestamp.Add(time.Minute),
			OpenPrice:  price,
			ClosePrice: price,
			HighPrice:  price,
			Volume:     volume,
		}
	}

	//Update current tempCandle with new trade data
	tempCandle.ClosePrice = price
	tempCandle.Volume += volume
	if price > tempCandle.HighPrice {
		tempCandle.HighPrice = price
	}
	if price < tempCandle.LowPrice {
		tempCandle.LowPrice = price
	}

	//Store the tempCandle for the symbol
	tempCandles[symbol] = tempCandle

	// write to the broadcast channel live ongoing channel
	broadcast <- &BroadCastMessage{
		UpdateType: Live,
		Candle:     tempCandle.toCandle(),
	}
}

// Send candle update to clients connected every 1 second at max, unless its a closed candle
func broadcastUpdates() {
	//set the broadcast interval to 1 sec
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var latestUpdate *BroadCastMessage

	for {
		select {
		//watch for new update from the broadcast channel
		case update := <-broadcast:
			// if the update is a closed candle, broadcast it immidiatly
			if update.UpdateType == Closes {
				// broadcast it
				broadcastToClient(update)
			} else {
				// replace temp update
				latestUpdate = update
			}

		case <-ticker.C:
			// broadcasr the latest update
			if latestUpdate != nil {
				// broadcast it
				broadcastToClient(latestUpdate)
			}
			latestUpdate = nil
		}
	}
}

// broadcast update to client
func broadcastToClient(update *BroadCastMessage) {
	//Marshal the update struct to json
	jsonUpdate, _ := json.Marshal(update)

	// send the update to all conncected clients subscribed to the symbols
	for clientConn, symbols := range clientConns {
		// if the client is subscribed to the symbols at the update
		if update.Candle.Symbol == symbols {
			//send the update to the client
			err := clientConn.WriteMessage(websocket.TextMessage, jsonUpdate)
			if err != nil {
				log.Println("Error sending message to client: ", err)
				clientConn.Close()
				delete(clientConns, clientConn)
			}
		}
	}
}
