package main

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

func (a *api) getUserFromContext(c echo.Context) (*domain.User, error) {
	user, ok := c.Get(userCtxKey).(*domain.User)
	if !ok {
		return nil, errors.New("user session not found")
	}
	return user, nil
}

func (a *api) getSelfHandler(c echo.Context) error {
	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	}

	return c.JSON(http.StatusOK, user)
}

func (a *api) getUsersHandler(c echo.Context) error {
	users, err := a.services.Users.GetMany(c.Request().Context())

	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, users)
}

func (a *api) searchUserHandler(c echo.Context) error {
	id := c.QueryParam("id")
	validId, err := uuid.Parse(id)
	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid user id")
	}
	user, err := a.services.Users.GetById(c.Request().Context(), validId)

	if err != nil {
		a.notFoundLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, domain.SearchUserResult{
		JoinedOn: user.CreatedAt,
		Id:       user.ID,
		FullName: user.FullName,
	})
}
