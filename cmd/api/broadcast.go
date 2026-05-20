package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	"github.com/myselfBZ/satjade-backend/internal/ws/events"
)

func (a *api) broadcastStatusChange(ctx context.Context, userId uuid.UUID, status string) error {
	friends, err := a.services.Friends.GetManyByUser(ctx, userId)
	if err != nil {
		return err
	}

	for _, f := range friends {
		client, ok := a.wsClients.Get(f.FriendId.String())
		if ok {
			client.WriteEvent(events.ServerSentEventPayload{
				Type: events.UserStatusChangeType,
				Body: &events.UserStatusChange{
					UserId: userId.String(),
					Status: status,
				},
			})

		}
	}

	return nil

}




func (a *api) notifyFriendRequest(ctx context.Context, request domain.FriendshipRequest) error {
	client, ok := a.wsClients.Get(request.ToId.String())

	if ok {
		// TODO please fucking figure out how to return an error
		client.WriteEvent(events.ServerSentEventPayload{
			Type: events.FriendRequestType,
			Body: &events.FriendRequest{
				FriendshipRequest: request,
			},
		})
	}

	return nil
}
