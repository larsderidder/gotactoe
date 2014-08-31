package tictactoe

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	decisionInterval = time.Second * 3 / 2
	timeBetweenGames = time.Second * 5
)

// Incoming votes
type voteMsg struct {
	Coord
	Player string
}

var VoteInput = make(chan []byte)

var board *Board

func GetBoard() *Board {
	return board
}

// Collect votes and play the game!
func PlayGoTacToe() {
	board = NewBoard()
	votes := make(map[Coord]int)
	decisionTimer := time.After(decisionInterval)
	for {
		select {
		case input := <-VoteInput:
			var vote voteMsg
			err := json.Unmarshal(input, &vote)
			if err != nil {
				// Invalid vote data, who cares
				continue
			}
			log.Printf("Received vote %d,%d from %s", vote.X, vote.Y, vote.Player)
			if vote.Player == fmt.Sprint(board.Turn) {
				coord := Coord{vote.X, vote.Y}
				// Only process votes for empty fields
				if board.Fields[coord] == EMPTY {
					votes[coord] += 1
				}
			} else {
				log.Println("Ignoring vote, not your turn!")
			}
		case <-decisionTimer:
			decide(votes)
			// New channel to remove votes while deciding
			VoteInput = make(chan []byte)
			votes = make(map[Coord]int)
			decisionTimer = time.After(decisionInterval)
		}
	}
}

func decide(votes map[Coord]int) {
	log.Printf("Votes: %+v", votes)

	votesByCount, max_count := getVotesByCount(votes)
	var decision Coord
	if len(votes) == 0 {
		log.Println("No decision made! Randomly playing.")
		decision = board.RandomMove()
	} else {
		decision = votesByCount[max_count][rand.Intn(len(votesByCount[max_count]))]
		log.Printf("Decided on %d,%d", decision.X, decision.Y)
	}
	board.Play(decision.X, decision.Y)
	Mh.Boards <- board

	log.Printf("New board: %v\n", board)
	outcome := board.Winner()
	if outcome != NONE {
		log.Printf("We have an outcome, and it is %s!", fmt.Sprint(outcome))
		Mh.Outcomes <- outcome
		time.Sleep(timeBetweenGames)
		board = NewBoard()
		Mh.Boards <- board
	}
}

func getVotesByCount(votes map[Coord]int) (map[int][]Coord, int) {
	voteByCount := make(map[int][]Coord)
	max_count := 0
	for vote, count := range votes {
		list := voteByCount[count]
		if list != nil {
			voteByCount[count] = append(voteByCount[count], vote)
		} else {
			voteByCount[count] = []Coord{vote}
		}
		if count > max_count {
			max_count = count
		}
	}
	return voteByCount, max_count
}
