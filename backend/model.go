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
