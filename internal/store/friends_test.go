package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

func TestFriendsStore_Create(t *testing.T) {
	pool := newTestPool(t)
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tx.Rollback(ctx) })

	queries := db.New(tx) // pass tx instead of pool
	store := &friendsStore{pool: pool, queries: queries}

	user1 := seedUser(t, ctx, queries)
	user2 := seedUser(t, ctx, queries)

	friend, err := store.Create(ctx, &CreateFriendParams{
		User1:  user1,
		User2:  user2,
		SelfId: user1,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if friend.FriendId != user2 {
		t.Errorf("got friend %v, want %v", friend.FriendId, user2)
	}


	// sequential operation for duplicate-test
	// cause this shit can't be ran seperately, you know


	_, err = store.Create(ctx, &CreateFriendParams{
		User1: user1,
		User2: user2,

		SelfId: user1,
	})

	if err != domain.ErrFriendsAlreadyExist{
		t.Fatalf("got '%v', want '%v'", err, domain.ErrFriendsAlreadyExist)
	}
}




func seedUser(t *testing.T, ctx context.Context, q *db.Queries) uuid.UUID {
	t.Helper()
	row, err := q.CreateUser(ctx, db.CreateUserParams{
		Role:         domain.ROLE_STUDENT,
		FullName:     "Test User",
		Email:        uuid.NewString() + "@test.com", // unique email
		PasswordHash: "hashed",
	})
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return row.ID.Bytes
}
