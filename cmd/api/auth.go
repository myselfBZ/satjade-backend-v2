package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	auth_service "github.com/myselfBZ/satjade-backend/internal/services/auth"
)


func (a *api) loginHandler(c echo.Context) error {
	var payload auth_service.LoginParams

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "malformed request payload")
	}

	if err := a.validator.Struct(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, formatValidationError(err))
	}

	resp, err := a.services.Auth.Login(c.Request().Context(), &payload)

	if err != nil {
		switch err {
		case domain.ErrRecordNotFound:
			a.notFoundLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case domain.ErrInvalidCredentials:
			a.unauthorizedLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusUnauthorized, err)
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (a *api) createAccountHandler(c echo.Context) error {
	var payload auth_service.CreateAccountParams

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "malformed request payload")
	}

	if err := a.validator.Struct(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, formatValidationError(err))
	}

	resp, err := a.services.Auth.CreateAccount(c.Request().Context(), &payload)

	if err != nil {
		switch err {
		case domain.ErrDuplicateEmail:
			a.conflictLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.JSON(http.StatusOK, resp)
}
