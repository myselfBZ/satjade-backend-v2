package store

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

type friendsStore struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

type CreateFriendParams struct {
	User1 uuid.UUID `json:"user1" validate:"required"`
	User2 uuid.UUID `json:"user2" validate:"required"`

	SelfId uuid.UUID
}

type CreateFriendRequestParams struct {
	ToId    uuid.UUID `json:"-"`
	FromId  uuid.UUID `json:"-"`
	Message string    `json:"message,omitempty"`
}

type AcceptFrienshipRequest struct {
	SelfId              uuid.UUID
	FriendshipRequestId uuid.UUID
}

// TODO: better error handling
// unexpected errors for now
func (s *friendsStore) AcceptFrienshipRequest(ctx context.Context,  requestId uuid.UUID) (domain.Friendship ,error) {
	var result domain.Friendship

	if err :=  withTx(s.pool, ctx, func(tx pgx.Tx) error {
		txConn := s.queries.WithTx(tx)

		 friendship, err := s.accept(ctx, txConn, requestId) 

		 if err != nil {
			return err
		}

		result = friendship
		return nil
	}); err != nil {
		return result, err
	}

	return result, nil
}

func (s *friendsStore) accept(ctx context.Context, tx *db.Queries, requestId uuid.UUID) (domain.Friendship ,error) {
	req, err := tx.DeleteFriendshipRequest(ctx, pgtype.UUID{Bytes: requestId, Valid: true})

	u1, u2 := req.ToID, req.FromID

	if u1.String() > u2.String() {
		u1, u2 = u2, u1
	}

	row, err := tx.CreateFriend(ctx, db.CreateFriendParams{
		User1: u1,
		User2: u2,
	})

	if err != nil {
		return domain.Friendship{}, err
	}

	return domain.Friendship{
		Id: row.ID.Bytes,
		User1: row.User1.Bytes,
		User2: row.User2.Bytes,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (s *friendsStore) DeleteFriendshipRequest(ctx context.Context, id uuid.UUID) error {
	_, err := s.queries.DeleteFriendshipRequest(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return domain.ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *friendsStore) GetFriendshipRequests(ctx context.Context, userId uuid.UUID) ([]domain.FriendshipRequest, error) {
	rows, err := s.queries.GetFriendshipRequestsByUser(ctx, pgtype.UUID{Bytes: userId, Valid: true})
	if err != nil {
		return nil, err
	}
	requests := make([]domain.FriendshipRequest, len(rows))

	for i, row := range rows {
		requests[i] = domain.FriendshipRequest{
			Id: row.ID.Bytes,
			From: struct {
				Id       uuid.UUID `json:"id"`
				FullName string    `json:"full_name"`
			}{
				Id:       row.ID.Bytes,
				FullName: row.FullName,
			},
			ToId:      row.ToID.Bytes,
			CreatedAt: row.CreatedAt.Time,
			Message:   row.Message.String,
		}
	}

	return requests, nil
}
func (s *friendsStore) CreateFriendRequest(ctx context.Context, params *CreateFriendRequestParams) (domain.FriendshipRequest ,error) {
	row, err := s.queries.CreateFriendshipRequest(ctx, db.CreateFriendshipRequestParams{
		ToID:    pgtype.UUID{Bytes: params.ToId, Valid: true},
		FromID:  pgtype.UUID{Bytes: params.FromId, Valid: true},
		Message: pgtype.Text{String: params.Message, Valid: params.Message != ""},
	})

	if err != nil {

		if strings.Contains(err.Error(), "no_self_request") {
			return domain.FriendshipRequest{}, domain.ErrSelfCantBeFriend
		}

		return domain.FriendshipRequest{}, err
	}

	return domain.FriendshipRequest{
		Id: row.ID.Bytes,
		From: struct{Id uuid.UUID "json:\"id\""; FullName string "json:\"full_name\""}{
			Id: row.FromID.Bytes,
			FullName: row.SenderFullName,
		},
		ToId: row.ToID.Bytes,
		CreatedAt: row.CreatedAt.Time,
		Message: row.Message.String,
	}, nil
}

func (s *friendsStore) GetManyByUser(ctx context.Context, userId uuid.UUID) ([]domain.Friend, error) {
	rows, err := s.queries.GetFriendsByUser(ctx, pgtype.UUID{Bytes: userId, Valid: true})

	if err != nil {
		return nil, err
	}

	users := make([]domain.Friend, len(rows))

	for i, row := range rows {
		users[i] = domain.Friend{
			FullName:     row.FullName,
			FriedsSince:  row.CreatedAt.Time,
			FriendId:     row.FriendID.Bytes,
			FriendshipId: row.FriendshipID.Bytes,
		}
	}

	return users, nil
}

func (s *friendsStore) Create(ctx context.Context, params *CreateFriendParams) (domain.Friend, error) {
	u1, u2 := params.User1, params.User2

	if u1.String() > u2.String() {
		u1, u2 = u2, u1
	}

	row, err := s.queries.CreateFriend(ctx, db.CreateFriendParams{
		User1: pgtype.UUID{Bytes: u1, Valid: true},
		User2: pgtype.UUID{Bytes: u2, Valid: true},
	})

	if err != nil {
		switch {
		case
			strings.Contains(err.Error(), "unique_friendship"),
			strings.Contains(err.Error(), "enforce_user_order"):

			return domain.Friend{}, domain.ErrFriendsAlreadyExist
		case strings.Contains(err.Error(), "no_self_friendship"):
			return domain.Friend{}, domain.ErrSelfCantBeFriend
		}
	}

	friendId := params.User2

	if params.User2 == params.SelfId {
		friendId = params.User1
	}

	return domain.Friend{
		FriendId:     friendId,
		FriendshipId: row.ID.Bytes,
		FriedsSince:  row.CreatedAt.Time,
	}, nil

}

func (s *friendsStore) Delete(ctx context.Context, friendshipId uuid.UUID) error {
	err := s.queries.DeleteFriend(ctx, pgtype.UUID{Bytes: friendshipId, Valid: true})
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return domain.ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}
