package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	questions_service "github.com/myselfBZ/satjade-backend/internal/services/questions"
)

func (a *api) getQuestionDistributionHandler(c echo.Context) error {

	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	}

	distribution, err := a.services.Questions.GetDistribution(c.Request().Context(), user.ID)

	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, distribution)
}

func (a *api) filterQuestionIdsHandler(c echo.Context) error {
	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}

	var payload questions_service.FilterParams

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "malformed request payload")
	}

	payload.UserID = user.ID

	ids, err := a.services.Questions.FilterIDs(c.Request().Context(), &payload)

	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ids)
}

func (a *api) getQuestionHandler(c echo.Context) error {
	id := c.Param("questionId")

	validId, err := uuid.Parse(id)
	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid uuid")
	}

	q, err := a.services.Questions.GetById(c.Request().Context(), validId)

	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			a.notFoundLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, q)
}

func (a *api) linkQuestionToModule(c echo.Context) error {
	moduleId, err := a.getUUIDFromParam("moduleId", c)

	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid module id")
	}

	var payload questions_service.LinkToModuleParams
	payload.ModuleId = moduleId

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "malformed request payload")
	}

	if err := a.validator.Struct(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, formatValidationError(err))
	}

	if err := a.services.Questions.LinkToModule(c.Request().Context(), &payload); err != nil {
		switch err {
		case domain.ErrQuestionNotFound, domain.ErrModuleNotFound:
			a.notFoundLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.NoContent(http.StatusCreated)
}

func (a *api) createQuestionToModule(c echo.Context) error {
	moduleId, err := a.getUUIDFromParam("moduleId", c)

	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid module id")
	}

	var payload questions_service.CreateToModuleParams
	payload.ModuleId = moduleId

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "malformed request payload")
	}

	if err := a.validator.Struct(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, formatValidationError(err))
	}

	if err := a.services.Questions.CreateToModule(c.Request().Context(), &payload); err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}

func (a *api) createQuestionHandler(c echo.Context) error {
	var payload questions_service.CreateQuestParams

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "malformed request payload")
	}

	if err := validateQuestionPayload(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	if _, err := a.services.Questions.Create(c.Request().Context(), &payload); err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}

func (a *api) createQuestionAttemptHandler(c echo.Context) error {
	var payload questions_service.CreateAttemptParams

	user, err := a.getUserFromContext(c)

	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return err
	}

	payload.UserID = user.ID

	payload.QuestionID, err = a.getUUIDFromParam("questionId", c)

	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid question id")
	}

	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "malformed request payload")
	}

	if err := a.validator.Struct(&payload); err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, formatValidationError(err))
	}

	attempt, err := a.services.Questions.CreateAttempt(c.Request().Context(), &payload)
	if err != nil {

		if err == domain.ErrNoCorrectChoice {
			a.badRequestLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusCreated, attempt)
}

func (a *api) getQuestionAttemptsByUserHandler(c echo.Context) error {
	userID, err := a.getUUIDFromParam("userId", c)
	if err != nil {
		a.badRequestLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	attempts, err := a.services.Questions.GetAttemptsByUser(c.Request().Context(), userID)
	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, attempts)
}

func validateQuestionPayload(p *questions_service.CreateQuestParams) error {
	if p.Type == "" {
		return fmt.Errorf("type is required")
	}
	if p.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if p.Skill == "" {
		return fmt.Errorf("skill is required")
	}
	if p.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if p.Difficulty == "" {
		return fmt.Errorf("difficulty is required")
	}
	if p.Explanation == "" {
		return fmt.Errorf("explanation is required")
	}

	switch p.Type {
	case "multiple_choice":
		if len(p.Choices) < 4 || len(p.Choices) > 4 {
			return fmt.Errorf("multiple choice questions must have exactly 4 choices")
		}

		correctCount := 0
		for _, choice := range p.Choices {
			if choice.Label == "" {
				return fmt.Errorf("choice label cannot be empty")
			}
			if choice.Body == "" {
				return fmt.Errorf("choice body cannot be empty")
			}

			if choice.IsCorrect {
				correctCount++
			}
		}

		if correctCount != 1 {
			return fmt.Errorf("multiple choice must have exactly one correct answer (found %d)", correctCount)
		}

	case "open_response":
		if p.OpenAnswerKey == nil {
			return fmt.Errorf("open_answer_key is required for open_response type")
		}
		if p.OpenAnswerKey.ModelAnswer == "" {
			return fmt.Errorf("model answer cannot be empty")
		}

	default:
		return fmt.Errorf("unsupported question type: %s", p.Type)
	}

	return nil
}
