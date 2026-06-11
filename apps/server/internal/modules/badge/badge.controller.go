package badge

import (
	"fmt"
	"net/http"
	"peekaping/internal/utils"
	"strconv"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type Controller struct {
	service Service
	logger  *zap.SugaredLogger
}

func NewController(service Service, logger *zap.SugaredLogger) *Controller {
	return &Controller{
		service: service,
		logger:  logger.Named("[badge-controller]"),
	}
}

// parseQueryOptions parses badge options from query parameters
func (c *Controller) parseQueryOptions(e *echo.Context) *BadgeOptions {
	options := DefaultBadgeOptions()

	// Parse style
	if style := e.QueryParam("style"); style != "" {
		switch style {
		case "flat", "flat-square", "plastic", "for-the-badge", "social":
			options.Style = BadgeStyle(style)
		}
	}

	// Parse common options
	if color := e.QueryParam("color"); color != "" {
		options.Color = color
	}
	if labelColor := e.QueryParam("labelColor"); labelColor != "" {
		options.LabelColor = labelColor
	}

	// Parse status badge options
	if upLabel := e.QueryParam("upLabel"); upLabel != "" {
		options.UpLabel = upLabel
	}
	if downLabel := e.QueryParam("downLabel"); downLabel != "" {
		options.DownLabel = downLabel
	}
	if upColor := e.QueryParam("upColor"); upColor != "" {
		options.UpColor = upColor
	}
	if downColor := e.QueryParam("downColor"); downColor != "" {
		options.DownColor = downColor
	}

	// Parse text customization options
	if labelPrefix := e.QueryParam("labelPrefix"); labelPrefix != "" {
		options.LabelPrefix = labelPrefix
	}
	if label := e.QueryParam("label"); label != "" {
		options.Label = label
	}
	if labelSuffix := e.QueryParam("labelSuffix"); labelSuffix != "" {
		options.LabelSuffix = labelSuffix
	}
	if prefix := e.QueryParam("prefix"); prefix != "" {
		options.Prefix = prefix
	}
	if suffix := e.QueryParam("suffix"); suffix != "" {
		options.Suffix = suffix
	}

	// Parse certificate expiry options
	if warnDays := e.QueryParam("warnDays"); warnDays != "" {
		if days, err := strconv.Atoi(warnDays); err == nil && days > 0 {
			options.WarnDays = days
		}
	}
	if downDays := e.QueryParam("downDays"); downDays != "" {
		if days, err := strconv.Atoi(downDays); err == nil && days >= 0 {
			options.DownDays = days
		}
	}

	return options
}

// @Router		/badge/{monitorId}/status [get]
// @Summary		Get status badge
// @Tags			Badges
// @Produce		image/svg+xml
// @Param			monitorId	path	string	true	"Monitor ID"
// @Param			style		query	string	false	"Badge style (flat, flat-square, plastic, for-the-badge, social)"
// @Param			upLabel		query	string	false	"Label when monitor is up"
// @Param			downLabel	query	string	false	"Label when monitor is down"
// @Param			upColor		query	string	false	"Color when monitor is up"
// @Param			downColor	query	string	false	"Color when monitor is down"
// @Success		200	{string}	string	"SVG badge"
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) GetStatusBadge(e *echo.Context) error {
	monitorID := e.Param("monitorId")
	if monitorID == "" {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Monitor ID is required"))
	}

	// Check if monitor is public (published on a status page)
	isPublic, err := c.service.IsMonitorPublic(e.Request().Context(), monitorID)
	if err != nil {
		c.logger.Errorw("Failed to check if monitor is public", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if !isPublic {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Monitor not found or not public"))
	}

	options := c.parseQueryOptions(e)

	svg, err := c.service.GenerateStatusBadge(e.Request().Context(), monitorID, options)
	if err != nil {
		c.logger.Errorw("Failed to generate status badge", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Failed to generate badge"))
	}

	e.Response().Header().Set("Content-Type", "image/svg+xml")
	e.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return e.String(http.StatusOK, svg)
}

// @Router		/badge/{monitorId}/uptime/{duration} [get]
// @Summary		Get uptime badge
// @Tags			Badges
// @Produce		image/svg+xml
// @Param			monitorId	path	string	true	"Monitor ID"
// @Param			duration	path	int		true	"Duration in hours (24, 720, 2160)"
// @Param			style		query	string	false	"Badge style"
// @Param			label		query	string	false	"Custom label"
// @Param			suffix		query	string	false	"Value suffix"
// @Param			color		query	string	false	"Badge color"
// @Success		200	{string}	string	"SVG badge"
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) GetUptimeBadge(e *echo.Context) error {
	monitorID := e.Param("monitorId")
	if monitorID == "" {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Monitor ID is required"))
	}

	durationStr := e.Param("duration")
	duration, err := strconv.Atoi(durationStr)
	fmt.Println("duration", duration)
	if err != nil || duration <= 0 {
		duration = 24 // Default to 24 hours
	}

	// Check if monitor is public
	isPublic, err := c.service.IsMonitorPublic(e.Request().Context(), monitorID)
	if err != nil {
		c.logger.Errorw("Failed to check if monitor is public", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if !isPublic {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Monitor not found or not public"))
	}

	options := c.parseQueryOptions(e)

	svg, err := c.service.GenerateUptimeBadge(e.Request().Context(), monitorID, duration, options)
	if err != nil {
		c.logger.Errorw("Failed to generate uptime badge", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Failed to generate badge"))
	}

	e.Response().Header().Set("Content-Type", "image/svg+xml")
	e.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return e.String(http.StatusOK, svg)
}

// @Router		/badge/{monitorId}/ping/{duration} [get]
// @Summary		Get ping badge
// @Tags			Badges
// @Produce		image/svg+xml
// @Param			monitorId	path	string	true	"Monitor ID"
// @Param			duration	path	int		true	"Duration in hours"
// @Param			style		query	string	false	"Badge style"
// @Param			label		query	string	false	"Custom label"
// @Param			suffix		query	string	false	"Value suffix"
// @Param			color		query	string	false	"Badge color"
// @Success		200	{string}	string	"SVG badge"
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) GetPingBadge(e *echo.Context) error {
	monitorID := e.Param("monitorId")
	if monitorID == "" {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Monitor ID is required"))
	}

	durationStr := e.Param("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		duration = 24 // Default to 24 hours
	}

	// Check if monitor is public
	isPublic, err := c.service.IsMonitorPublic(e.Request().Context(), monitorID)
	if err != nil {
		c.logger.Errorw("Failed to check if monitor is public", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if !isPublic {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Monitor not found or not public"))
	}

	options := c.parseQueryOptions(e)

	svg, err := c.service.GeneratePingBadge(e.Request().Context(), monitorID, duration, options)
	if err != nil {
		c.logger.Errorw("Failed to generate ping badge", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Failed to generate badge"))
	}

	e.Response().Header().Set("Content-Type", "image/svg+xml")
	e.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return e.String(http.StatusOK, svg)
}

// @Router		/badge/{monitorId}/cert-exp [get]
// @Summary		Get certificate expiry badge
// @Tags			Badges
// @Produce		image/svg+xml
// @Param			monitorId	path	string	true	"Monitor ID"
// @Param			style		query	string	false	"Badge style"
// @Param			label		query	string	false	"Custom label"
// @Param			warnDays	query	int		false	"Days before expiry to show warning"
// @Param			downDays	query	int		false	"Days before expiry to show as down"
// @Success		200	{string}	string	"SVG badge"
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) GetCertExpBadge(e *echo.Context) error {
	monitorID := e.Param("monitorId")
	if monitorID == "" {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Monitor ID is required"))
	}

	// Check if monitor is public
	isPublic, err := c.service.IsMonitorPublic(e.Request().Context(), monitorID)
	if err != nil {
		c.logger.Errorw("Failed to check if monitor is public", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if !isPublic {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Monitor not found or not public"))
	}

	options := c.parseQueryOptions(e)

	svg, err := c.service.GenerateCertExpBadge(e.Request().Context(), monitorID, options)
	if err != nil {
		c.logger.Errorw("Failed to generate cert-exp badge", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Failed to generate badge"))
	}

	e.Response().Header().Set("Content-Type", "image/svg+xml")
	e.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return e.String(http.StatusOK, svg)
}

// @Router		/badge/{monitorId}/response [get]
// @Summary		Get response time badge
// @Tags			Badges
// @Produce		image/svg+xml
// @Param			monitorId	path	string	true	"Monitor ID"
// @Param			style		query	string	false	"Badge style"
// @Param			label		query	string	false	"Custom label"
// @Param			suffix		query	string	false	"Value suffix"
// @Param			color		query	string	false	"Badge color"
// @Success		200	{string}	string	"SVG badge"
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) GetResponseBadge(e *echo.Context) error {
	monitorID := e.Param("monitorId")
	if monitorID == "" {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Monitor ID is required"))
	}

	// Check if monitor is public
	isPublic, err := c.service.IsMonitorPublic(e.Request().Context(), monitorID)
	if err != nil {
		c.logger.Errorw("Failed to check if monitor is public", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if !isPublic {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Monitor not found or not public"))
	}

	options := c.parseQueryOptions(e)

	svg, err := c.service.GenerateResponseBadge(e.Request().Context(), monitorID, options)
	if err != nil {
		c.logger.Errorw("Failed to generate response badge", "error", err, "monitorID", monitorID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Failed to generate badge"))
	}

	e.Response().Header().Set("Content-Type", "image/svg+xml")
	e.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return e.String(http.StatusOK, svg)
}
