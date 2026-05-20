package challenge

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)


var(
	ErrChallengeNotFound = errors.New("challenge not found")
	ErrReciepientIdMismatch = errors.New("recepient id did not match")
)


func NewManager() Manager {
	return &manager{
		pendingChallengeRequests: make(map[string]ChallengeRequest),
		challengeMu: sync.Mutex{},
		expired: make(chan ChallengeRequest),
	}
}

type ChallengeRequest struct {
	Id 	   string `json:"id"`
	FromId string `json:"from_id"`
	ToId   string `json:"to_id"`
}



type Manager interface {
	Create(toId string, fromId string) (string, error)
	Accept(id string, recepientId string) (ChallengeRequest, error)

	ExpiredCh() <- chan ChallengeRequest
}

type manager struct {
	pendingChallengeRequests map[string]ChallengeRequest
	challengeMu sync.Mutex

	expired   chan ChallengeRequest
}

func (s *manager) ExpiredCh() <- chan ChallengeRequest  {
	return  s.expired
}


func (s *manager) Accept(id string, recepientId string) (ChallengeRequest, error) {
	s.challengeMu.Lock()
	defer s.challengeMu.Unlock()

	r, ok := s.pendingChallengeRequests[id]

	if !ok {
		return ChallengeRequest{}, ErrChallengeNotFound
	}

	if r.ToId != recepientId {
		return ChallengeRequest{}, ErrReciepientIdMismatch
	}

	delete(s.pendingChallengeRequests, r.Id)
	return r, nil
}

func (s *manager) Create(toId string, fromId string) (string, error) {
	id, err := uuid.NewUUID()

	if err != nil {
		return "", err
	}

	s.challengeMu.Lock()
	defer s.challengeMu.Unlock()

	s.pendingChallengeRequests[id.String()] = ChallengeRequest{
		Id: id.String(),
		FromId: fromId,
		ToId: toId,
	}


	go func(id string) {
		<- time.After(time.Second * 5)

		s.challengeMu.Lock()
		r, ok := s.pendingChallengeRequests[id] 

		if ok {
			delete(s.pendingChallengeRequests, r.Id)
			s.challengeMu.Unlock()
			s.expired <- r
			return
		}
		s.challengeMu.Unlock()

	}(id.String())

	return id.String(), nil
}
