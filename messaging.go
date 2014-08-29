package main

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

type MessageType string

// Constants for message types
const (
	BOARD   MessageType = "board"
	OUTCOME             = "outcome"
)

// Global settings for connections
const (
	pongDeadline = 60 * time.Second

	pingPeriod = pongDeadline * 8 / 10

	writeDeadline = 5 * time.Second
)

type Serializer interface {
	Json() []byte
}

type Message struct {
	Type MessageType
}

type OutcomeMsg struct {
	Message
	Outcome string
}

func (msg *OutcomeMsg) Json() []byte {
	val, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return val
}

type BoardMsg struct {
	Message
	Fields [][]Field
	Turn   string
}

func (msg *BoardMsg) Json() []byte {
	val, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return val
}

type messageHandler struct {
	boards   chan *Board
	outcomes chan Outcome
}

var mh = messageHandler{
	boards:   make(chan *Board),
	outcomes: make(chan Outcome),
}

func (mh *messageHandler) handle() {
	for {
		select {
		case board := <-mh.boards:
			msg := NewBoardMsg(board)
			h.broadcast <- &msg
		case outcome := <-mh.outcomes:
			outcomeMsg := OutcomeMsg{Message{OUTCOME}, OutcomeToString(outcome)}
			h.broadcast <- &outcomeMsg
		}
	}
}

// Factory function to create a message for a board.
func NewBoardMsg(board *Board) BoardMsg {
	fields := board.FieldsList()
	return BoardMsg{Message: Message{Type: BOARD}, Fields: fields, Turn: PlayerToString(board.turn)}
}

type hub struct {
	connections map[*websocket.Conn]bool
	broadcast   chan Serializer
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
}

var h = hub{
	connections: make(map[*websocket.Conn]bool),
	broadcast:   make(chan Serializer),
	register:    make(chan *websocket.Conn),
	unregister:  make(chan *websocket.Conn),
}

func (h *hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.connections[conn] = true
		case conn := <-h.unregister:
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				conn.Close()
			}
		case msg := <-h.broadcast:
			for conn := range h.connections {
				go sendMsg(conn, websocket.TextMessage, msg.Json())
			}
		}
	}
}

func waitForMessages(conn *websocket.Conn) {
	h.register <- conn
	defer func() {
		h.unregister <- conn
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

// Low level function to send a message to a connection
func sendMsg(conn *websocket.Conn, msgType int, msg []byte) error {
	conn.SetWriteDeadline(time.Now().Add(writeDeadline))
	return conn.WriteMessage(msgType, msg)
}
