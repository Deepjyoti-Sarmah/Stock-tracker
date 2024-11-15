package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	symbols = []string{"AAPL", "AMZN"}
	// symbols := []string{"AAPL", "AMZN", "BINANCE:BTCUSDT", "IC MARKETS:1"}
	//
	// Broadcast messages to all connected clients
	broadcast = make(chan *BroadCastMessage)

	//map of all ongoing live candles for each symbol
	tempCandles = make(map[string]*TempCandle)

	//
	mu sync.Mutex
)

func main() {
	//Env config
	env := EnvConnfig()

	// db connection
	db := DBConnection(env)

	// connect to finnhub websocket
	finnhubWSConn := connectToFinnhub(env)
	defer finnhubWSConn.Close()

	// handle finnhub incomming message
	go handleFinnhubMessages(finnhubWSConn, db)
	//
	// Broadcast candle update to all clients connected
	//
	//--- Endpoints---
	//Connect to the websocket
	//fetch all past candle for all of the symbol
	//fetch all past candle for specific symbol
	//
	//Serve the endpoints

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
	finnhubMessage := FinnhubMessage{}

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
				UpdateType: Close,
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
	if price < tempCandle.HighPrice {
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
