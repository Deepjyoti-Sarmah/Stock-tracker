package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

// Message represent the structure of the websocket Message
type Message struct {
	MessageType int
	Data        []byte
}

func main() {
	// Connect to the remote ws
	u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/ws"}
	fmt.Printf("Connecting to %s\n", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	//Channels for mannaging messages
	send := make(chan Message)
	done := make(chan struct{})

	//Goroutines for reading messages
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read: ", err)
				return
			}

			fmt.Printf("Received: %s\n", message)
		}
	}()

	//Goroutines for sending messages
	go func() {
		for {
			select {
			case msg := <-send:
				//write that to the websocket connection
				err := conn.WriteMessage(msg.MessageType, msg.Data)
				if err != nil {
					log.Println("write: ", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	//Read input from terminal and send it to the websocket server
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type something...")
	for scanner.Scan() {
		text := scanner.Text()
		// send the text to the Channels
		send <- Message{websocket.TextMessage, []byte(text)}
	}

	if err := scanner.Err(); err != nil {
		log.Println("scanner err: ", err)
	}

}
