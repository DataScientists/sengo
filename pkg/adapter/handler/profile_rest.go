package handler

import (
	"errors"
	"net/http"
	"strings"

	"sheng-go-backend/pkg/adapter/controller"
	"sheng-go-backend/pkg/entity/model"
	"sheng-go-backend/pkg/infrastructure/external/rapidapi"
	routerhandler "sheng-go-backend/pkg/infrastructure/router/handler"

	"github.com/labstack/echo/v4"
)

// ProfileRESTHandler exposes REST endpoints around profile fetching.
type ProfileRESTHandler struct {
	profileEntry controller.ProfileEntry
}

// NewProfileRESTHandler creates a ProfileRESTHandler.
func NewProfileRESTHandler(profileEntry controller.ProfileEntry) *ProfileRESTHandler {
	return &ProfileRESTHandler{profileEntry: profileEntry}
}

type fetchProfileRequest struct {
	LinkedinURL string `json:"linkedinUrl"`
	Gender      string `json:"gender"`
}

// Fetch handles POST /api/profiles/fetch.
//
// It accepts a LinkedIn profile URL and gender, then returns the stored
// profile data, fetching it on demand when it has not been fetched yet.
func (h *ProfileRESTHandler) Fetch(c echo.Context) error {
	var req fetchProfileRequest
	if err := c.Bind(&req); err != nil {
		return routerhandler.HandleError(c, model.NewInvalidParamError(req))
	}

	if strings.TrimSpace(req.LinkedinURL) == "" {
		return routerhandler.HandleError(c, model.NewInvalidParamError("linkedinUrl is required"))
	}

	var gender *string
	if g := strings.TrimSpace(req.Gender); g != "" {
		gender = &g
	}

	profile, err := h.profileEntry.FetchProfileByURL(c.Request().Context(), req.LinkedinURL, gender)
	if err != nil {
		return routerhandler.HandleError(c, toRESTError(err))
	}

	return c.JSON(http.StatusOK, profile)
}

// toRESTError normalises errors from the fetch flow into model errors that the
// shared HandleError helper understands.
func toRESTError(err error) error {
	// Already a model error (validation, db, etc.).
	var coded interface{ Code() string }
	if errors.As(err, &coded) {
		return err
	}

	// Profile not found upstream at RapidAPI.
	var notFound *rapidapi.NotFoundError
	if errors.As(err, &notFound) {
		return model.NewNotFoundError(err, notFound.URN)
	}

	return model.NewInternalServerError(err)
}
