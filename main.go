package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
