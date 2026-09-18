package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthResponse is returned by the health-check endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
	Version   string `json:"version"`
}

// HealthHandler handles GET /health — used by load-balancers and monitoring tools.
func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "gts-backend",
		Version:   "1.0.0",
	})
}
