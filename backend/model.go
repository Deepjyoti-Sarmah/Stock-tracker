package main

import "time"

// Candle struct represents a single OHLC (Open, High, Low, Close) candle
type Candle struct {
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	High      float64   `json:"hign"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Timestamp time.Time `json:"timestamp"`
}

// TempCandle represents an item from the temp canlde slice building the candles
type TempCandle struct {
	Symbol     string
	OpenTime   time.Time
	CloseTime  time.Time
	OpenPrice  float64
	ClosePrice float64
	HighPrice  float64
	LowPrice   float64
	Volume     float64
}

// Structure of the data comming from the finnhub websocket api
type FinnhubMessage struct {
	Data []TradeData `json:"data"`
	Type string      `json:"type"` //ping or trade
}

type TradeData struct {
	Close     []string `json:"c"`
	Price     float64  `json:"p"`
	Symbol    string   `json:"s"`
	Timestamp int64    `json:"t"`
	Volume    int      `json:"v"`
}

// Data to write to client connected
type BroadCastMessage struct {
	UpdateType UpdateType `json:"updateType"` // "live" : "closes"
	Candle     *Candle    `json:"candle"`
}

type UpdateType string

const (
	Live  UpdateType = "live"   //realtime ongoing candle
	Closes UpdateType = "closes" //past candle already Closed
)

// converts a tempCandle to candle
func (tx *TempCandle) toCandle() *Candle {
	return &Candle{
		Symbol:    tx.Symbol,
		Open:      tx.OpenPrice,
		Close:     tx.ClosePrice,
		High:      tx.HighPrice,
		Low:       tx.LowPrice,
		Timestamp: tx.CloseTime,
	}
}
