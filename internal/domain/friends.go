package domain

import (
	"time"

	"github.com/google/uuid"
)

type Friend struct {
	FullName     string    `json:"full_name"`
	FriendId     uuid.UUID `json:"friend_id"`
	FriendshipId uuid.UUID `json:"friendship_id"`
	FriedsSince  time.Time `json:"friends_since"`
	IsOnline     bool      `json:"is_online"`
}

type FriendshipRequest struct {
	Id   uuid.UUID `json:"id"`
	From struct {
		Id       uuid.UUID `json:"id"`
		FullName string    `json:"full_name"`
	} `json:"from"`
	ToId      uuid.UUID `json:"to_id"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}



// This type is kind of a duck tape for now
type Friendship struct {
	Id        uuid.UUID `json:"id"`
	User1     uuid.UUID `json:"user_1"`
	User2     uuid.UUID `json:"user_2"`
	CreatedAt time.Time `json:"created_at"`

	OnlineStatus bool `json:"online_status"`
}
