package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	practices_service "github.com/myselfBZ/satjade-backend/internal/services/practices"
)



func (a *api) getFullTest(c echo.Context) error {
	id, err := a.getUUIDFromParam("practiceId", c)
	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid practice id")
	}

	p, err := a.services.Practices.GetFullTest(c.Request().Context(), id)

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

	return c.JSON(http.StatusOK, p)
}

func (a *api) createPracticeHandler(c echo.Context) error {
	var payload practices_service.CreatePracticeParams
	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "malformed request payload")
	}
	if err := a.validator.Struct(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, formatValidationError(err))
	}
	practice, err := a.services.Practices.Create(c.Request().Context(), payload.Title)
	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusCreated, practice)
}

func (a *api) getPublishedPracticeHandler(c echo.Context) error {
	publishedId, err := a.getUUIDFromParam("publishedId", c)

	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid practice id")
	}

	practice, err := a.services.Practices.GetPublishedById(c.Request().Context(), publishedId)

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

	return c.JSON(http.StatusOK, practice)
}

func (a *api) getPublishedPracticesPreviewsHandler(c echo.Context) error {
	practices, err := a.services.Practices.GetPublishedPreviews(c.Request().Context())

	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, practices)
}


func (a *api) pulishPracticeHandler(c echo.Context) error {
	practiceId, err := a.getUUIDFromParam("practiceId", c)
	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid practice id")
	}

	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}
	
	err = a.services.Practices.Publish(c.Request().Context(), &practices_service.PublishParams{
		PublishedBy: user.ID,
		PracticeId: practiceId,
	})

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

	return c.NoContent(http.StatusCreated)
}

func (a *api) getPracticePreviewsHandler(c echo.Context) error {
	practices, err := a.services.Practices.GetPreviews(c.Request().Context())
	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, practices)
}
