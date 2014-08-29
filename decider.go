package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"
)

var voteInput = make(chan []byte)

const (
	decisionInterval = time.Second * 1
	timeBetweenGames = time.Second * 5
)

// Incoming votes
type voteMsg struct {
	Coord
	Player string
}

func CollectVotes() {
	votes := make(map[Coord]int)
	decisionTimer := time.After(decisionInterval)
	deciding := false
	newRound := make(chan bool)
	for {
		select {
		case input := <-voteInput:
			if deciding {
				// Drain votes while deciding.
				continue
			}
			var vote voteMsg
			err := json.Unmarshal(input, &vote)
			if err != nil {
				// Invalid vote data, who cares
				continue
			}
			log.Printf("Received vote %d,%d from %s", vote.X, vote.Y, vote.Player)
			if vote.Player == PlayerToString(board.turn) {
				coord := Coord{vote.X, vote.Y}
				// Only process votes for empty fields
				if board.fields[coord] == EMPTY {
					votes[coord] += 1
				}
			} else {
				log.Println("Ignoring vote, not your turn!")
			}
		case <-decisionTimer:
			// Do this in a goroutine so we can drain votes while deciding.
			go decide(newRound, votes)
			deciding = true
		case <-newRound:
			deciding = false
			votes = make(map[Coord]int)
			decisionTimer = time.After(decisionInterval)
		}
	}
}

func decide(newRound chan bool, votes map[Coord]int) {
	log.Println("Votes: ", votes)
	defer func() {
		newRound <- true
	}()

	votesByCount, max_count := getVotesByCount(votes)
	var decision Coord
	if len(votes) == 0 {
		log.Println("No decision made! Randomly playing.")
		decision = board.RandomMove()
	} else {
		decision = votesByCount[max_count][rand.Intn(len(votesByCount[max_count]))]
		log.Println("Decided on %d,%d", decision.X, decision.Y)
	}
	board.Play(decision.X, decision.Y)
	mh.Boards <- &board

	log.Println("New board: " + board.repr())
	outcome := board.Winner()
	if outcome != NONE {
		log.Printf("We have an outcome, and it is %s!", OutcomeToString(outcome))
		mh.Outcomes <- outcome
		time.Sleep(timeBetweenGames)
		board = NewBoard()
		mh.Boards <- &board
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
