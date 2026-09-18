package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tahakhan/gts_backend/models"
	"github.com/tahakhan/gts_backend/services"
)

// DownloadHandler handles GET /api/v1/download/:id
// It looks up the audio metadata by ID and streams the file back to the client.
type DownloadHandler struct {
	extractor *services.ExtractionService
}

// NewDownloadHandler creates a DownloadHandler wired to the given ExtractionService.
func NewDownloadHandler(extractor *services.ExtractionService) *DownloadHandler {
	return &DownloadHandler{extractor: extractor}
}

// Handle is the Echo handler function for the download endpoint.
//
//	GET /api/v1/download/:id
//
// The handler resolves the audio file path from the in-memory store and uses
// http.ServeFile to stream the content back, which automatically handles:
//   - Range requests (resumable downloads)
//   - Content-Type detection
//   - ETags / Last-Modified caching headers
func (h *DownloadHandler) Handle(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_id",
			Message: "path parameter ':id' is required",
		})
	}

	meta, found := h.extractor.GetByID(id)
	if !found {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "no audio file found with the given ID; it may have expired",
		})
	}

	// Set a descriptive Content-Disposition so the Flutter app (or browser)
	// knows the suggested filename.
	filename := meta.Title + "." + meta.Format
	c.Response().Header().Set(
		"Content-Disposition",
		`attachment; filename="`+filename+`"`,
	)

	// http.ServeFile handles range requests, ETags, and Content-Type automatically.
	http.ServeFile(c.Response(), c.Request(), meta.FilePath)
	return nil
}
