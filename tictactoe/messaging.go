package tictactoe

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

type MessageType string

// Constants for message types
const (
	BOARD    MessageType = "board"
	OUTCOME              = "outcome"
	REGISTER             = "register"
	STATS                = "stats"
)

// Global settings for connections
const (
	writeDeadline = time.Second * 5
	statsInterval = time.Second * 10
)

type Serializer interface {
	Serialize() []byte
}

type Message struct {
	Type MessageType
}

type OutcomeMsg struct {
	Message
	Outcome string
}

func (msg *OutcomeMsg) Serialize() []byte {
	return MustJson(msg)
}

type BoardMsg struct {
	Message
	Fields [][]Field
	Turn   string
}

func (msg *BoardMsg) Serialize() []byte {
	return MustJson(msg)
}

type RegisterMsg struct {
	Message
	Player string
}

func (msg *RegisterMsg) Serialize() []byte {
	return MustJson(msg)
}

type StatsMsg struct {
	Message
	XPlayers int
	OPlayers int
}

func (msg *StatsMsg) Serialize() []byte {
	return MustJson(msg)
}

func (h *hub) countPlayers(p Player) int {
	nr := 0
	for conn := range h.connections {
		if p == h.connections[conn] {
			nr++
		}
	}
	return nr
}

func NewStatsMsg() *StatsMsg {
	return &StatsMsg{Message{STATS}, Hub.countPlayers(CROSS), Hub.countPlayers(CIRCLE)}
}

func MustJson(msg interface{}) []byte {
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

var Mh = messageHandler{
	Boards:   make(chan *Board),
	Outcomes: make(chan Outcome),
}

func (Mh *messageHandler) handle() {
	statsTimer := time.After(statsInterval)
	for {
		select {
		case board := <-Mh.Boards:
			Hub.broadcast <- NewBoardMsg(board)
		case outcome := <-Mh.Outcomes:
			Hub.broadcast <- &OutcomeMsg{Message{OUTCOME}, fmt.Sprint(outcome)}
		case <-statsTimer:
			Hub.broadcast <- NewStatsMsg()
			statsTimer = time.After(statsInterval)
		}
	}
}

// Factory function to create a message for a board.
func NewBoardMsg(board *Board) *BoardMsg {
	fields := board.FieldsList()
	return &BoardMsg{Message: Message{Type: BOARD}, Fields: fields, Turn: fmt.Sprint(board.Turn)}
}

type hub struct {
	connections map[*websocket.Conn]Player
	broadcast   chan Serializer
	Register    chan *websocket.Conn
	Unregister  chan *websocket.Conn
}

var Hub = hub{
	connections: make(map[*websocket.Conn]Player),
	broadcast:   make(chan Serializer),
	Register:    make(chan *websocket.Conn),
	Unregister:  make(chan *websocket.Conn),
}

func (h *hub) run() {
	for {
		select {
		case conn := <-h.Register:
			h.connections[conn] = newPlayer()
			go SendMsg(conn, websocket.TextMessage, NewBoardMsg(board).Serialize())
			msg := RegisterMsg{Message{REGISTER}, fmt.Sprint(h.connections[conn])}
			go SendMsg(conn, websocket.TextMessage, msg.Serialize())
		case conn := <-h.Unregister:
			// Check if connection exist, only if so continue
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				conn.Close()
			}
		case msg := <-h.broadcast:
			for conn := range h.connections {
				go SendMsg(conn, websocket.TextMessage, msg.Serialize())
			}
		}
	}
}

func newPlayer() Player {
	xPlayers, oPlayers := Hub.countPlayers(CROSS), Hub.countPlayers(CIRCLE)
	if xPlayers > oPlayers {
		return CIRCLE
	}
	if xPlayers < oPlayers {
		return CROSS
	}
	return Players[rand.Intn(len(Players))]
}

// Low level function to send a message to a connection
func SendMsg(conn *websocket.Conn, msgType int, msg []byte) error {
	conn.SetWriteDeadline(time.Now().Add(writeDeadline))
	err := conn.WriteMessage(msgType, msg)
	if err != nil {
		Hub.Unregister <- conn
	}
	return err
}
