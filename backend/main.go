package main

func main() {
	//Env config
	env := EnvConnfig()

	// db connection
	db := DBConnection(env)

	// connect to finnhub websocket
	//
	// handle finnhub incomming message
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
