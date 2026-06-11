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

type TechnologyController struct {
	technologyService *service.TechnologyService
}

func NewTechnologyController(technologyService *service.TechnologyService) *TechnologyController {
	return &TechnologyController{technologyService: technologyService}
}

func (c *TechnologyController) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetTechnologies(w, r)
	case http.MethodPost:
		c.CreateTechnology(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *TechnologyController) GetTechnologies(w http.ResponseWriter, r *http.Request) {
	technologies, err := c.technologyService.GetTechnologies(r.Context(), model.TechnologyFilter{
		Query:         r.URL.Query().Get("q"),
		Category:      r.URL.Query().Get("category"),
		OwnerTeam:     r.URL.Query().Get("owner_team"),
		Status:        r.URL.Query().Get("status"),
		RiskLevel:     r.URL.Query().Get("risk_level"),
		AdoptionStage: r.URL.Query().Get("adoption_stage"),
		Tag:           r.URL.Query().Get("tag"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list technologies")
		return
	}
	writeJSON(w, http.StatusOK, technologies)
}

func (c *TechnologyController) Show(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/technologies/")
	switch r.Method {
	case http.MethodGet:
		c.GetTechnology(w, r, id)
	case http.MethodPut:
		c.UpdateTechnology(w, r, id)
	case http.MethodDelete:
		c.DeleteTechnology(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *TechnologyController) GetTechnology(w http.ResponseWriter, r *http.Request, id string) {
	technology, err := c.technologyService.GetTechnologyByID(r.Context(), id)
	if errors.Is(err, repository.ErrTechnologyNotFound) {
		writeError(w, http.StatusNotFound, "technology not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, technology)
}

func (c *TechnologyController) CreateTechnology(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var request model.CreateTechnologyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	technology, err := c.technologyService.CreateTechnology(r.Context(), request)
	if errors.Is(err, service.ErrInvalidTechnology) {
		writeError(w, http.StatusBadRequest, "name, category, version and owner_team are required")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create technology")
		return
	}
	writeJSON(w, http.StatusCreated, technology)
}

func (c *TechnologyController) UpdateTechnology(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()
	var request model.UpdateTechnologyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	technology, err := c.technologyService.UpdateTechnology(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidTechnology) {
		writeError(w, http.StatusBadRequest, "name, category, version and owner_team are required")
		return
	}
	if errors.Is(err, repository.ErrTechnologyNotFound) {
		writeError(w, http.StatusNotFound, "technology not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update technology")
		return
	}
	writeJSON(w, http.StatusOK, technology)
}

func (c *TechnologyController) DeleteTechnology(w http.ResponseWriter, r *http.Request, id string) {
	err := c.technologyService.DeleteTechnology(r.Context(), id)
	if errors.Is(err, service.ErrInvalidTechnology) {
		writeError(w, http.StatusBadRequest, "technology id is required")
		return
	}
	if errors.Is(err, repository.ErrTechnologyNotFound) {
		writeError(w, http.StatusNotFound, "technology not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete technology")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
