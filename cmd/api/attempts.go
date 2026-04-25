package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	practiceattempt_service "github.com/myselfBZ/satjade-backend/internal/services/practice-attempt"
)

func (a *api) createPracticeAttemptHandler(c echo.Context) error {
	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}

	var payload practiceattempt_service.CreatePracticeAttemptParams
	payload.UserId = user.ID

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "malformed request payload")
	}

	if err := a.validator.Struct(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, formatValidationError(err))
	}

	if err := a.services.PracticeAttempt.Create(c.Request().Context(), &payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "unexpected error")
	}

	return c.NoContent(http.StatusCreated)
}

func (a *api) getPracticeAttemptByIdHandler(c echo.Context) error {
	id, err := a.getUUIDFromParam("attemptId", c)

	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, formatValidationError(err))
	}

	attempt, err := a.services.PracticeAttempt.GetById(c.Request().Context(), id)

	if err != nil {

		switch err {
		case domain.ErrRecordNotFound:
			a.notFoundLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.JSON(http.StatusOK, attempt)
}

func (a *api) getPracticeAttemptPreviewsHandler(c echo.Context) error {
	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}
	previews, err := a.services.PracticeAttempt.GetPreviewsByUser(c.Request().Context(), user.ID)

	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, previews)
}





