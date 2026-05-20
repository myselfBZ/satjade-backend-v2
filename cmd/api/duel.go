package main

import (
	"github.com/google/uuid"
	"sync"
)

type duel struct {
	Id             uuid.UUID
	user1, user2   string
	correctAnswers map[string]int
	mu             sync.Mutex

	doneUsers 	   map[string]bool

	exitCh chan struct{}
}

func (d *duel) getWinnerId() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	user1Score := d.correctAnswers[d.user1]
	user2Score := d.correctAnswers[d.user2]

	if user1Score > user2Score {
		return d.user1
	} else if user1Score < user2Score {
		return d.user2
	}

	return ""
}

func (d *duel) isUserDone(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.doneUsers[id]	
	return ok
}


// potential race. but ok for now. no modifications after creation
func (d *duel) getPeer(selfId string) string {
	if selfId == d.user1 {
		return d.user2
	}
	return d.user1
}

func (d *duel) doneUsersLen() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	return len(d.doneUsers)
}

func (d *duel) setDoneUser(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.doneUsers[id] = true
}

func (d *duel) incrementScore(userId string) {
	d.correctAnswers[userId]++
}

func newDuelMap() *duelMap {
	return &duelMap{
		duels: make(map[string]*duel),
		mu:    sync.Mutex{},
	}
}

type duelMap struct {
	duels map[string]*duel
	mu    sync.Mutex
}

func (d *duelMap) Get(id string) (*duel, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	dl, ok := d.duels[id]
	return dl, ok
}

func (d *duelMap) Set(id string, dl *duel) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.duels[id] = dl
}

func (d *duelMap) Del(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.duels, id)
}
