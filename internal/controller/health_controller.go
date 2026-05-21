package controller

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type HealthController struct {
	db *mongo.Database
}

func NewHealthController(db *mongo.Database) *HealthController {
	return &HealthController{db: db}
}

func (c *HealthController) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (c *HealthController) Ready(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := c.db.Client().Ping(ctx, readpref.Primary()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not_ready",
			"mongo":  "down",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
		"mongo":  "ok",
	})
}
