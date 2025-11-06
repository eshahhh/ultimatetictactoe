package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

func main() {
	var playerName string
	if len(os.Args) > 1 {
		playerName = os.Args[1]
	}

	serverURL := "ws://localhost:8080/ws"
	if playerName != "" {
		serverURL += "?name=" + playerName
	}

	fmt.Println("Connecting to Ultimate Tic-Tac-Toe Matchmaking Server")

	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to server! Entering matchmaking queue")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}
			fmt.Printf("\nServer: %s\n", string(message))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter command (help for commands): ")

	go func() {
		for scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())

			if input == "quit" || input == "exit" {
				fmt.Println("Disconnecting")
				close(done)
				return
			}

			if input != "" {
				err := conn.WriteMessage(websocket.TextMessage, []byte(input))
				if err != nil {
					log.Println("Error sending message:", err)
					return
				}
			}

			fmt.Print("Enter command: ")
		}
	}()

	select {
	case <-done:
		return
	case <-interrupt:
		fmt.Println("\nReceived interrupt signal. Closing connection")

		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Error sending close message:", err)
			return
		}

		return
	}
}
