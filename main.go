package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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
	h.Register <- conn
	defer func() {
		h.Unregister <- conn
	}()
	conn.SetReadDeadline(time.Now().Add(pongDeadline))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongDeadline)); return nil })

	// Start separate goroutine for keepalive, as conn.ReadMessage() blocks
	go func(conn *websocket.Conn) {
		// Regularly sends ping messages, to make sure the connection is still out there.
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			conn.Close()
		}()
		for {
			<-ticker.C
			if err := sendMsg(conn, websocket.PingMessage, []byte{}); err != nil {
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
		voteInput <- msg
	}
}

func boardHandler(writer http.ResponseWriter, request *http.Request) {
	msg := NewBoardMsg(&board)
	val, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(writer, string(val))
}

func main() {
	go h.run()
	go mh.handle()
	board = NewBoard()
	go CollectVotes()
	// Set delimiters for templates to not conflict with Angular
	http.Handle("/", http.FileServer(http.Dir("templates")))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/board", boardHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	port := 8080
	log.Printf("We are listening (on port %d)", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
