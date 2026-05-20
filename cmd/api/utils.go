package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func slowDown(secs int) {
	time.Sleep(time.Duration(secs) * time.Second)
}

func PrintRoutes(e *echo.Echo) {
	fmt.Println("\n--- Registered Routes ---")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

	// Header
	fmt.Fprintln(w, "METHOD\tPATH\tHANDLER")
	fmt.Fprintln(w, "------\t----\t-------")

	for _, route := range e.Routes() {
		if strings.Contains(route.Method, "echo_route_not_found") {
			continue
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", route.Method, route.Path, route.Name)
	}

	w.Flush()
	fmt.Println("-------------------------")
}

func (a *api) getUUIDFromParam(name string, c echo.Context) (uuid.UUID, error) {
	id := c.Param(name)

	validID, err := uuid.Parse(id)

	if err != nil {
		return uuid.UUID{}, err
	}

	return validID, nil
}

func formatValidationError(err error) string {
	errors := ""

	valErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return "Internal validation error"
	}

	for _, f := range valErrors {
		var msg string

		switch f.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is a required field. ", f.Field())
		case "email":
			msg = fmt.Sprintf("%s must be a valid email address. ", f.Field())
		case "gte":
			msg = fmt.Sprintf("%s must be at least %s. ", f.Field(), f.Param())
		case "lte":
			msg = fmt.Sprintf("%s cannot be greater than %s. ", f.Field(), f.Param())
		case "min":
			msg = fmt.Sprintf("%s must be at least %s characters. ", f.Field(), f.Param())
		case "max":
			msg = fmt.Sprintf("%s cannot exceed %s characters. ", f.Field(), f.Param())
		default:
			msg = fmt.Sprintf("%s is not valid input\n", f.Field())
		}

		errors += msg
	}

	return errors
}
