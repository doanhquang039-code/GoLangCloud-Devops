package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"hr-cloud-service/internal/model"
	"hr-cloud-service/internal/repository"
	"hr-cloud-service/internal/service"
)

type ActivityController struct {
	activityService *service.ActivityService
}

func NewActivityController(activityService *service.ActivityService) *ActivityController {
	return &ActivityController{activityService: activityService}
}

func (c *ActivityController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetActivities(w, r)
	case http.MethodPost:
		c.CreateActivity(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *ActivityController) GetActivities(w http.ResponseWriter, r *http.Request) {
	activities, err := c.activityService.GetActivities(r.Context(), model.ActivityFilter{
		Query:         r.URL.Query().Get("q"),
		Type:          r.URL.Query().Get("type"),
		Action:        r.URL.Query().Get("action"),
		Status:        r.URL.Query().Get("status"),
		Actor:         r.URL.Query().Get("actor"),
		ResourceType:  r.URL.Query().Get("resource_type"),
		ResourceID:    r.URL.Query().Get("resource_id"),
		ApplicationID: r.URL.Query().Get("application_id"),
		OwnerTeam:     r.URL.Query().Get("owner_team"),
		Tag:           r.URL.Query().Get("tag"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list activities")
		return
	}
	writeJSON(w, http.StatusOK, activities)
}

func (c *ActivityController) Show(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/activities/")
	switch r.Method {
	case http.MethodGet:
		c.GetActivity(w, r, id)
	case http.MethodPut:
		c.UpdateActivity(w, r, id)
	case http.MethodDelete:
		c.DeleteActivity(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *ActivityController) GetActivity(w http.ResponseWriter, r *http.Request, id string) {
	activity, err := c.activityService.GetActivityByID(r.Context(), id)
	if errors.Is(err, repository.ErrActivityNotFound) {
		writeError(w, http.StatusNotFound, "activity not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, activity)
}

func (c *ActivityController) CreateActivity(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var request model.CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	activity, err := c.activityService.CreateActivity(r.Context(), request)
	if errors.Is(err, service.ErrInvalidActivity) {
		writeError(w, http.StatusBadRequest, "type, action, actor, resource_type, resource_id and summary are required")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create activity")
		return
	}
	writeJSON(w, http.StatusCreated, activity)
}

func (c *ActivityController) UpdateActivity(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()
	var request model.UpdateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	activity, err := c.activityService.UpdateActivity(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidActivity) {
		writeError(w, http.StatusBadRequest, "type, action, actor, resource_type, resource_id and summary are required")
		return
	}
	if errors.Is(err, repository.ErrActivityNotFound) {
		writeError(w, http.StatusNotFound, "activity not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update activity")
		return
	}
	writeJSON(w, http.StatusOK, activity)
}

func (c *ActivityController) DeleteActivity(w http.ResponseWriter, r *http.Request, id string) {
	err := c.activityService.DeleteActivity(r.Context(), id)
	if errors.Is(err, service.ErrInvalidActivity) {
		writeError(w, http.StatusBadRequest, "activity id is required")
		return
	}
	if errors.Is(err, repository.ErrActivityNotFound) {
		writeError(w, http.StatusNotFound, "activity not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete activity")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
