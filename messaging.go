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
	Boards   chan *Board
	Outcomes chan Outcome
}

var mh = messageHandler{
	Boards:   make(chan *Board),
	Outcomes: make(chan Outcome),
}

func (mh *messageHandler) handle() {
	for {
		select {
		case board := <-mh.Boards:
			msg := NewBoardMsg(board)
			h.broadcast <- &msg
		case outcome := <-mh.Outcomes:
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
	Register    chan *websocket.Conn
	Unregister  chan *websocket.Conn
}

var h = hub{
	connections: make(map[*websocket.Conn]bool),
	broadcast:   make(chan Serializer),
	Register:    make(chan *websocket.Conn),
	Unregister:  make(chan *websocket.Conn),
}

func (h *hub) run() {
	for {
		select {
		case conn := <-h.Register:
			h.connections[conn] = true
		case conn := <-h.Unregister:
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

// Low level function to send a message to a connection
func sendMsg(conn *websocket.Conn, msgType int, msg []byte) error {
	conn.SetWriteDeadline(time.Now().Add(writeDeadline))
	return conn.WriteMessage(msgType, msg)
}
