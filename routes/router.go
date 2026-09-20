// Package routes wires Echo routes to their handler implementations.
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tahakhan/gts_backend/handlers"
	"github.com/tahakhan/gts_backend/services"
)

// Register attaches all application routes to the given Echo instance.
// It follows the versioned API prefix /api/v1 for resource endpoints.
func Register(e *echo.Echo, extractor *services.ExtractionService) {
	extractHandler := handlers.NewExtractHandler(extractor)
	downloadHandler := handlers.NewDownloadHandler(extractor)

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// ── Health ────────────────────────────────────────────────────────────────
	e.GET("/health", handlers.HealthHandler)

	// ── API v1 ────────────────────────────────────────────────────────────────
	v1 := e.Group("/api/v1")

	// POST /api/v1/extract
	// Body: { "url": "<youtube-url>", "format": "mp3" }
	// Returns: ExtractionResponse with a download_url
	v1.POST("/extract", extractHandler.Handle)

	// GET /api/v1/download/:id
	// Streams the extracted audio file back to the client.
	v1.GET("/download/:id", downloadHandler.Handle)
}
