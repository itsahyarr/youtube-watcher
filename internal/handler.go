package internal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ScrapeRequest struct {
	URL   string  `json:"url" binding:"required"`
	Proxy *string `json:"proxy"`
}

type Handler struct {
	cfg *Config
	svc *Service
}

func NewHandler(cfg *Config, svc *Service) *Handler {
	return &Handler{cfg: cfg, svc: svc}
}

// ScrapeYouTube godoc
//
//	@summary		Scrape a YouTube video
//	@description	Opens a YouTube URL in a browser, clicks the play button, and logs the result to MongoDB.
//	@tags			scrape
//	@accept			json
//	@produce		json
//	@param			headless	query		bool			false	"Run browser in headless mode"
//	@param			request		body		ScrapeRequest	true	"Scrape request body"
//	@success		200			{object}	SuccessResponse
//	@failure		400			{object}	ErrorResponse
//	@failure		500			{object}	ErrorResponse
//	@router			/api/v1/scrape/youtube/play [post]
func (h *Handler) ScrapeYouTube(c *gin.Context) {
	var req ScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Status:  "BAD_REQUEST",
			Success: false,
			Errors: map[string]interface{}{
				"message": "url is required and must be a valid YouTube URL",
			},
		})
		return
	}

	headless := parseHeadlessParam(c.Query("headless"), h.cfg.RodHeadless)

	if !IsValidYouTubeURL(req.URL) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Status:  "BAD_REQUEST",
			Success: false,
			Errors: map[string]interface{}{
				"message": "url is not a valid YouTube URL",
			},
		})
		return
	}

	proxyURL := ""
	if req.Proxy != nil {
		proxyURL = *req.Proxy
		if err := ValidateProxy(proxyURL); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    400,
				Status:  "BAD_REQUEST",
				Success: false,
				Errors: map[string]interface{}{
					"message": fmt.Sprintf("invalid proxy: %v", err),
				},
			})
			return
		}
	}

	logEntry, err := h.svc.ExecuteScrape(c.Request.Context(), req.URL, proxyURL, headless)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Status:  "INTERNAL_SERVER_ERROR",
			Success: false,
			Errors: map[string]interface{}{
				"message": err.Error(),
			},
		})
		return
	}

	if logEntry.Status != "SUCCESS" {
		errMsg := logEntry.Message
		if logEntry.Error != nil {
			errMsg = *logEntry.Error
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Status:  "INTERNAL_SERVER_ERROR",
			Success: false,
			Errors: map[string]interface{}{
				"message": errMsg,
			},
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    200,
		Status:  "OK",
		Success: true,
		Data: ScrapeData{
			LogID:   logEntry.ID.Hex(),
			URL:     logEntry.URL,
			Action:  logEntry.Action,
			Result:  "SUCCESS",
			Message: logEntry.Message,
		},
	})
}

func parseHeadlessParam(param string, defaultVal bool) bool {
	switch param {
	case "true":
		return true
	case "false":
		return false
	default:
		return defaultVal
	}
}
