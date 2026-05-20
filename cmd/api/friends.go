package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	friends_service "github.com/myselfBZ/satjade-backend/internal/services/friends"
	"github.com/myselfBZ/satjade-backend/internal/store"
	"github.com/myselfBZ/satjade-backend/internal/ws/events"
)

// POST /users/{user_id}/friends
func (a *api) sendFriendshipRequestHandler(c echo.Context) error {
	toId, err := a.getUUIDFromParam("user_id", c)

	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid user id")

	}

	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}

	payload := friends_service.CreateFriendshipRequest{
		CreateFriendRequestParams: &store.CreateFriendRequestParams{},
	}

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "malformed json input")
	}

	payload.ToId = toId
	payload.FromId = user.ID

	// TODO
	request ,err := a.services.Friends.CreateFriendshipRequest(c.Request().Context(), &payload) 
	if err != nil {

		if err == domain.ErrSelfCantBeFriend {
			a.badRequestLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
		}

		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error happened")
	}

	a.notifyFriendRequest(c.Request().Context(), request)


	return c.NoContent(http.StatusOK)
}

// POST /friendship/requests/{request_id}
func (a *api) acceptFriendshipRequestHandler(c echo.Context) error {
	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}

	requestId, err := a.getUUIDFromParam("request_id", c)
	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid request id")
	}

	// TODOOOOOOOOO better errors?
	result, err := a.services.Friends.AcceptFriendshipRequest(c.Request().Context(), requestId)

	if err != nil {
		switch err {
		case domain.ErrRecordNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "friendship request not found")
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error happened")
		}
	}

	friendId := result.User1

	if user.ID == friendId {
		friendId = result.User2
	}

	client, ok := a.wsClients.Get(friendId.String())

	result.OnlineStatus = ok

	if ok {
		client.WriteEvent(events.ServerSentEventPayload{
			Type: events.NewFriendType,
			Body: &events.NewFriend{
				FullName:     user.FullName,
				FriendId:     user.ID,
				FriendsSince: result.CreatedAt,
				FriendshipId: result.Id,
				IsOnline:     true,
			},
		})
	}

	return c.JSON(http.StatusOK, result)
}

// DELETE /friendship/requests/{request_id}
func (a *api) rejectFriendshipRequestHandler(c echo.Context) error {
	requestId, err := a.getUUIDFromParam("request_id", c)
	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid request id")
	}

	// TODOOOOOO better errorss pleaseee
	if err := a.services.Friends.DeleteFriendshipRequest(c.Request().Context(), requestId); err != nil {
		switch err {
		case domain.ErrRecordNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "friendship request not found")
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error happened")
		}
	}
	return c.NoContent(http.StatusOK)
}

// GET /friendship/requests/
func (a *api) getFriendshipRequestsHandler(c echo.Context) error {
	user, err := a.getUserFromContext(c)
	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}
	requests, err := a.services.Friends.GetFriendshipRequests(c.Request().Context(), user.ID)
	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error happened")
	}
	return c.JSON(http.StatusOK, requests)
}

// GET /user/me/friends
func (a *api) getFriendsHandler(c echo.Context) error {
	user, err := a.getUserFromContext(c)
	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}
	friends, err := a.services.Friends.GetManyByUser(c.Request().Context(), user.ID)
	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error happened")
	}

	for i := range friends {
		_, ok := a.wsClients.Get(friends[i].FriendId.String())
		friends[i].IsOnline = ok
	}

	return c.JSON(http.StatusOK, friends)
}

// DELETE /friendship/:id
func (a *api) deleteFriendHandler(c echo.Context) error {
	friendshipId, err := a.getUUIDFromParam("id", c)
	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid friendship id")
	}
	if err := a.services.Friends.Delete(c.Request().Context(), friendshipId); err != nil {
		switch err {
		case domain.ErrRecordNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "friendship not found")
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error happened")
		}
	}
	return c.NoContent(http.StatusOK)
}
