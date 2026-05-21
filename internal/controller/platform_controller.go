package controller

import (
	"net/http"

	"hr-cloud-service/internal/service"
)

type PlatformController struct {
	platformService *service.PlatformService
}

func NewPlatformController(platformService *service.PlatformService) *PlatformController {
	return &PlatformController{platformService: platformService}
}

func (c *PlatformController) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	summary, err := c.platformService.GetSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not build platform summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
