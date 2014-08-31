/*
A small web application for collaboratively playing tic-tac-toe.

main.go contains the web handlers dealing with requests and websockets, and contains the main function.

messaging.go contains all message structs, plus a hub to manage connections.

tictactoe.go contains all information needed to play tictactoe, such as the board with its method and the Player type.

decider.go is where the actual game is played, and votes are collected for moves.
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/todayispotato/gotactoe/tictactoe"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	pongDeadline = time.Second * 5
	pingPeriod   = pongDeadline * 8 / 10
)

func wsHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(writer, "Method not allowed", 405)
		return
	}
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("We have a new connection, rejoice \\o/")
	waitForMessages(conn)
}

func waitForMessages(conn *websocket.Conn) {
	tictactoe.Hub.Register <- conn
	defer func() {
		tictactoe.Hub.Unregister <- conn
	}()
	conn.SetReadDeadline(time.Now().Add(pongDeadline))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongDeadline)); return nil })

	// Start separate goroutine for keepalive, as conn.ReadMessage() blocks
	go func(conn *websocket.Conn) {
		// Regularly sends ping messages, to make sure the connection is still out there.
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			tictactoe.Hub.Unregister <- conn
		}()
		for {
			<-ticker.C
			if err := tictactoe.SendMsg(conn, websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}(conn)

	// Wait for messages!
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		tictactoe.VoteInput <- msg
	}
}

// Handler for debugging purposes
func boardHandler(writer http.ResponseWriter, request *http.Request) {
	msg := tictactoe.NewBoardMsg(tictactoe.GetBoard())
	val, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(writer, string(val))
}

func main() {
	go tictactoe.PlayGoTacToe()
	http.Handle("/", http.FileServer(http.Dir("templates")))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/board", boardHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	port := os.Getenv("PORT")
	log.Printf("We are listening (on port %s)", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
