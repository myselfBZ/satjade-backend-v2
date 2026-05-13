package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

func (a *api) broadcastOfflineStatus(ctx context.Context, userId uuid.UUID) error {
	friends, err := a.services.Friends.GetManyByUser(ctx, userId)
	if err != nil {
		return err
	}

	for _, f := range friends {
		client, ok := a.wsClients.Get(f.FriendId.String())
		if ok {
			if err := client.writeEvent(ctx, serverSentEvent{
				Type: offlineStatusType,
				Body: &offlineStatus{
					UserId: userId.String(),
				},
			}); err != nil {
				return err
			}

		}
	}

	return nil

}


func (a *api) broadcastOnlineStatus(ctx context.Context, userId uuid.UUID) error {
	friends, err := a.services.Friends.GetManyByUser(ctx, userId)
	if err != nil {
		return err
	}

	for _, f := range friends {
		client, ok := a.wsClients.Get(f.FriendId.String())
		if ok {
			if err := client.writeEvent(ctx, serverSentEvent{
				Type: onlineStatusType,
				Body: &onlineStatus{
					UserId: userId.String(),
				},
			}); err != nil {
				return err
			}

		}
	}

	return nil

}




func (a *api) notifyFriendRequest(ctx context.Context, request domain.FriendshipRequest) error {
	client, ok := a.wsClients.Get(request.ToId.String())

	if ok {
		return client.writeEvent(ctx, serverSentEvent{
			Type: friendRequestType,
			Body: &friendRequest{
				FriendshipRequest: request,
			},
		})
	}

	return nil
}
