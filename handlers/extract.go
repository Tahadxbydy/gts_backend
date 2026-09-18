// Package handlers contains Echo HTTP handler functions.
package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tahakhan/gts_backend/models"
	"github.com/tahakhan/gts_backend/services"
)

// ExtractHandler handles POST /api/v1/extract
// It validates the request body, triggers yt-dlp extraction, and returns
// an ExtractionResponse with a download URL the client can use.
type ExtractHandler struct {
	extractor *services.ExtractionService
}

// NewExtractHandler creates an ExtractHandler wired to the given ExtractionService.
func NewExtractHandler(extractor *services.ExtractionService) *ExtractHandler {
	return &ExtractHandler{extractor: extractor}
}

// Handle is the Echo handler function for the extract endpoint.
//
//	POST /api/v1/extract
//	Content-Type: application/json
//	Body: { "url": "https://www.youtube.com/watch?v=...", "format": "mp3" }
func (h *ExtractHandler) Handle(c echo.Context) error {
	var req models.AudioRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "request body must be valid JSON with a 'url' field",
		})
	}

	if req.URL == "" {
		return c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{
			Error:   "missing_url",
			Message: "'url' field is required",
		})
	}

	// Perform the extraction (this call is blocking — consider moving to async + status polling
	// in a production version).
	meta, err := h.extractor.Extract(&req)
	if err != nil {
		c.Logger().Errorf("extraction failed for url=%s: %v", req.URL, err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "extraction_failed",
			Message: err.Error(),
		})
	}

	resp := models.ExtractionResponse{
		ID:          meta.ID,
		Title:       meta.Title,
		Format:      meta.Format,
		DownloadURL: fmt.Sprintf("/api/v1/download/%s", meta.ID),
		CreatedAt:   meta.CreatedAt,
	}

	return c.JSON(http.StatusCreated, resp)
}
