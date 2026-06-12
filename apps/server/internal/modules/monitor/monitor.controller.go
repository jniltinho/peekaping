package monitor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"peekaping/internal/modules/monitor_notification"
	"peekaping/internal/modules/monitor_tag"
	"peekaping/internal/modules/monitor_tls_info"
	"peekaping/internal/utils"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type MonitorController struct {
	monitorService             Service
	logger                     *zap.SugaredLogger
	monitorNotificationService monitor_notification.Service
	monitorTagService          monitor_tag.Service
	tlsInfoService             monitor_tls_info.Service
}

func NewMonitorController(
	monitorService Service,
	logger *zap.SugaredLogger,
	monitorNotificationService monitor_notification.Service,
	monitorTagService monitor_tag.Service,
	tlsInfoService monitor_tls_info.Service,
) *MonitorController {
	utils.Validate.RegisterStructValidation(CreateUpdateDtoStructLevelValidation, CreateUpdateDto{})

	return &MonitorController{
		monitorService,
		logger,
		monitorNotificationService,
		monitorTagService,
		tlsInfoService,
	}
}

// @Router		/monitors [get]
// @Summary		Get monitors
// @Tags			Monitors
// @Produce		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     q    query     string  false  "Search query"
// @Param     page query     int     false  "Page number" default(1)
// @Param     limit query    int     false  "Items per page" default(10)
// @Param     active query   bool    false  "Active status"
// @Param     status query   int     false  "Status"
// @Param     tag_ids query  string  false  "Comma-separated list of tag IDs to filter by"
// @Success		200	{object}	utils.ApiResponse[[]Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *MonitorController) FindAll(e *echo.Context) error {
	page, err := utils.GetQueryInt(e, "page", 0)
	if err != nil || page < 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid page parameter"))
	}

	limit, err := utils.GetQueryInt(e, "limit", 10)
	if err != nil || limit < 1 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid limit parameter"))
	}

	q := e.QueryParam("q")

	active, err := utils.GetQueryBool(e, "active")
	if err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid active parameter (must be true or false)"))
	}

	var statusPtr *int
	if statusStr := e.QueryParam("status"); statusStr != "" {
		statusVal, err := utils.GetQueryInt(e, "status", 0)
		if err != nil {
			return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid status parameter (must be int)"))
		}
		statusPtr = &statusVal
	}

	// Parse tag_ids parameter
	var tagIds []string
	if tagIdsStr := e.QueryParam("tag_ids"); tagIdsStr != "" {
		tagIds = strings.Split(tagIdsStr, ",")
		// Trim whitespace from each tag ID
		for i, tagId := range tagIds {
			tagIds[i] = strings.TrimSpace(tagId)
		}
		// Remove empty strings
		var validTagIds []string
		for _, tagId := range tagIds {
			if tagId != "" {
				validTagIds = append(validTagIds, tagId)
			}
		}
		tagIds = validTagIds
	}

	response, err := ic.monitorService.FindAll(e.Request().Context(), page, limit, q, active, statusPtr, tagIds)
	if err != nil {
		ic.logger.Errorw("Failed to fetch monitors", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", response))
}

// @Router		/monitors [post]
// @Summary		Create monitor
// @Tags			Monitors
// @Produce		json
// @Accept		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     body body   CreateUpdateDto  true  "Monitor object"
// @Success		201	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *MonitorController) Create(e *echo.Context) error {
	var monitor *CreateUpdateDto
	if err := e.Bind(&monitor); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := utils.Validate.Struct(monitor); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// Validate monitor type and config
	if err := ic.monitorService.ValidateMonitorConfig(monitor.Type, monitor.Config); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(fmt.Sprintf("Invalid monitor configuration: %v", err)))
	}

	createdMonitor, err := ic.monitorService.Create(e.Request().Context(), monitor)
	if err != nil {
		ic.logger.Errorw("Failed to create monitor", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	ic.logger.Infof("Created monitor: %+v\n", createdMonitor)

	// Handle multiple notification IDs
	if len(monitor.NotificationIds) > 0 {
		for _, notificationId := range monitor.NotificationIds {
			_, err = ic.monitorNotificationService.Create(e.Request().Context(), createdMonitor.ID, notificationId)
			if err != nil {
				ic.logger.Errorw("Failed to create monitor-notification record", "error", err)
				return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
			}
		}
	}

	// Handle multiple tag IDs
	if len(monitor.TagIds) > 0 {
		for _, tagId := range monitor.TagIds {
			_, err = ic.monitorTagService.Create(e.Request().Context(), createdMonitor.ID, tagId)
			if err != nil {
				ic.logger.Errorw("Failed to create monitor-tag record", "error", err)
				return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
			}
		}
	}

	return e.JSON(http.StatusCreated, utils.NewSuccessResponse("Monitor created successfully", createdMonitor))
}

// @Router		/monitors/{id} [get]
// @Summary		Get monitor by ID
// @Tags			Monitors
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Monitor ID"
// @Success		200	{object}	utils.ApiResponse[MonitorResponseDto]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *MonitorController) FindByID(e *echo.Context) error {
	id := e.Param("id")

	monitor, err := ic.monitorService.FindByID(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to fetch monitor", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	if monitor == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Monitor not found"))
	}

	// Fetch notification_ids
	notificationRels, err := ic.monitorNotificationService.FindByMonitorID(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to fetch monitor-notification relations", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	notificationIds := make([]string, 0, len(notificationRels))
	for _, rel := range notificationRels {
		notificationIds = append(notificationIds, rel.NotificationID)
	}

	// Fetch tag_ids
	tagRels, err := ic.monitorTagService.FindByMonitorID(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to fetch monitor-tag relations", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	tagIds := make([]string, 0, len(tagRels))
	for _, rel := range tagRels {
		tagIds = append(tagIds, rel.TagID)
	}

	// Compose response with notification_ids and tag_ids
	response := MonitorResponseDto{
		ID:              monitor.ID,
		Name:            monitor.Name,
		Interval:        monitor.Interval,
		Timeout:         monitor.Timeout,
		Type:            monitor.Type,
		Active:          monitor.Active,
		MaxRetries:      monitor.MaxRetries,
		RetryInterval:   monitor.RetryInterval,
		ResendInterval:  monitor.ResendInterval,
		Status:          int(monitor.Status),
		CreatedAt:       monitor.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       monitor.UpdatedAt.Format(time.RFC3339),
		NotificationIds: notificationIds,
		TagIds:          tagIds,
		ProxyId:         monitor.ProxyId,
		Config:          monitor.Config,
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", response))
}

// @Router		/monitors/{id} [put]
// @Summary		Update monitor
// @Tags			Monitors
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Monitor ID"
// @Param       monitor body     CreateUpdateDto  true  "Monitor object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *MonitorController) UpdateFull(e *echo.Context) error {
	id := e.Param("id")

	var monitor CreateUpdateDto
	if err := e.Bind(&monitor); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// validate
	if err := utils.Validate.Struct(monitor); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// Validate monitor type and config
	if err := ic.monitorService.ValidateMonitorConfig(monitor.Type, monitor.Config); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(fmt.Sprintf("Invalid monitor configuration: %v", err)))
	}

	updatedMonitor, err := ic.monitorService.UpdateFull(e.Request().Context(), id, &monitor)
	if err != nil {
		ic.logger.Errorw("Failed to update monitor", "error", err)
		if errors.Is(err, ErrMonitorNotFound) {
			return e.JSON(http.StatusNotFound, utils.NewFailResponse(err.Error()))
		}
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	// Delete all existing notification relations and create new ones
	err = ic.monitorNotificationService.DeleteByMonitorID(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to delete existing monitor-notification relations", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	// Create new notification relations
	for _, notificationId := range monitor.NotificationIds {
		_, err = ic.monitorNotificationService.Create(e.Request().Context(), id, notificationId)
		if err != nil {
			ic.logger.Errorw("Failed to create monitor-notification record", "error", err)
			return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		}
	}

	// Delete all existing tag relations and create new ones
	err = ic.monitorTagService.DeleteByMonitorID(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to delete existing monitor-tag relations", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	// Create new tag relations
	for _, tagId := range monitor.TagIds {
		_, err = ic.monitorTagService.Create(e.Request().Context(), id, tagId)
		if err != nil {
			ic.logger.Errorw("Failed to create monitor-tag record", "error", err)
			return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		}
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("Monitor updated successfully", updatedMonitor))
}

// @Router		/monitors/{id} [patch]
// @Summary		Update monitor
// @Tags			Monitors
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Monitor ID"
// @Param       monitor body     PartialUpdateDto  true  "Monitor object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *MonitorController) UpdatePartial(e *echo.Context) error {
	id := e.Param("id")

	var monitor PartialUpdateDto
	if err := e.Bind(&monitor); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// Validate monitor type and config if they are being updated
	if monitor.Type != nil && monitor.Config != nil {
		if err := ic.monitorService.ValidateMonitorConfig(*monitor.Type, *monitor.Config); err != nil {
			return e.JSON(http.StatusBadRequest, utils.NewFailResponse(fmt.Sprintf("Invalid monitor configuration: %v", err)))
		}
	}

	updatedMonitor, err := ic.monitorService.UpdatePartial(e.Request().Context(), id, &monitor, false)
	if err != nil {
		ic.logger.Errorw("Failed to update monitor", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	// Handle notification IDs if they are being updated
	if len(monitor.NotificationIds) > 0 {
		// Replace all monitor-notification relations in an optimized way
		existing, err := ic.monitorNotificationService.FindByMonitorID(e.Request().Context(), id)
		if err != nil {
			ic.logger.Errorw("Failed to fetch monitor-notification relations", "error", err)
			return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		}

		// Build sets for comparison
		existingMap := make(map[string]string) // notificationID -> relationID
		for _, rel := range existing {
			existingMap[rel.NotificationID] = rel.ID
		}
		newSet := make(map[string]struct{})
		for _, nid := range monitor.NotificationIds {
			newSet[nid] = struct{}{}
		}

		// Delete relations not in the new list
		for notificationID, relID := range existingMap {
			if _, found := newSet[notificationID]; !found {
				if err := ic.monitorNotificationService.Delete(e.Request().Context(), relID); err != nil {
					ic.logger.Warnw("Failed to delete monitor-notification relation", "error", err)
				}
			}
		}

		// Add new relations not already present
		for _, nid := range monitor.NotificationIds {
			if _, found := existingMap[nid]; !found {
				if _, err := ic.monitorNotificationService.Create(e.Request().Context(), id, nid); err != nil {
					ic.logger.Warnw("Failed to create monitor-notification relation", "error", err)
				}
			}
		}
	}

	// Handle tag IDs if they are being updated
	if len(monitor.TagIds) > 0 {
		// Replace all monitor-tag relations in an optimized way
		existing, err := ic.monitorTagService.FindByMonitorID(e.Request().Context(), id)
		if err != nil {
			ic.logger.Errorw("Failed to fetch monitor-tag relations", "error", err)
			return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		}

		// Build sets for comparison
		existingMap := make(map[string]string) // tagID -> relationID
		for _, rel := range existing {
			existingMap[rel.TagID] = rel.ID
		}
		newSet := make(map[string]struct{})
		for _, tid := range monitor.TagIds {
			newSet[tid] = struct{}{}
		}

		// Delete relations not in the new list
		for tagID, relID := range existingMap {
			if _, found := newSet[tagID]; !found {
				if err := ic.monitorTagService.Delete(e.Request().Context(), relID); err != nil {
					ic.logger.Warnw("Failed to delete monitor-tag relation", "error", err)
				}
			}
		}

		// Add new relations not already present
		for _, tid := range monitor.TagIds {
			if _, found := existingMap[tid]; !found {
				if _, err := ic.monitorTagService.Create(e.Request().Context(), id, tid); err != nil {
					ic.logger.Warnw("Failed to create monitor-tag relation", "error", err)
				}
			}
		}
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("Monitor updated successfully", updatedMonitor))
}

// @Router		/monitors/{id} [delete]
// @Summary		Delete monitor
// @Tags			Monitors
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Monitor ID"
// @Success		200	{object}	utils.ApiResponse[any]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *MonitorController) Delete(e *echo.Context) error {
	id := e.Param("id")

	err := ic.monitorService.Delete(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to delete monitor", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Monitor deleted successfully", nil))
}

// @Router	/monitors/{id}/heartbeats [get]
// @Summary	Get paginated heartbeats for a monitor
// @Tags		Monitors
// @Produce	json
// @Security BearerAuth
// @Param	id	path	string	true	"Monitor ID"
// @Param	limit	query	int	false	"Number of heartbeats per page (default 50)"
// @Param	page	query	int	false	"Page number (default 0)"
// @Param	important	query	bool	false	"Filter by important heartbeats only"
// @Param	reverse	query	bool	false	"Reverse the order of heartbeats"
// @Success	200	{object}	utils.ApiResponse[[]heartbeat.Model]
// @Failure	400	{object}	utils.APIError[any]
// @Failure	404	{object}	utils.APIError[any]
// @Failure	500	{object}	utils.APIError[any]
func (ic *MonitorController) FindByMonitorIDPaginated(e *echo.Context) error {
	id := e.Param("id")

	limit, err := utils.GetQueryInt(e, "limit", 50)
	if err != nil || limit < 1 || limit > 1000 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid limit parameter (1-1000)"))
	}

	page, err := utils.GetQueryInt(e, "page", 0)
	if err != nil || page < 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid page parameter (>=0)"))
	}

	var importantPtr *bool
	if e.QueryParam("important") != "" {
		importantPtr, err = utils.GetQueryBool(e, "important")
		if err != nil {
			return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid important parameter (must be true or false)"))
		}
	}

	reverse := false
	if e.QueryParam("reverse") != "" {
		reversePtr, err := utils.GetQueryBool(e, "reverse")
		if err != nil {
			return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid reverse parameter (must be true or false)"))
		}
		if reversePtr != nil {
			reverse = *reversePtr
		}
	}

	results, err := ic.monitorService.GetHeartbeats(e.Request().Context(), id, limit, page, importantPtr, reverse)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		ic.logger.Errorw("Failed to get heartbeats", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", results))
}

// @Router /monitors/{id}/stats/points [get]
// @Summary Get monitor stat points (ping/up/down) from stats tables
// @Tags Monitors
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Param since query string true "Start time (RFC3339)"
// @Param until query string false "End time (RFC3339, default now)"
// @Param granularity query string false "Granularity (minute, hour, day)"
// @Success 200 {object} utils.ApiResponse[StatPointsSummaryDto]
// @Failure 400 {object} utils.APIError[any]
// @Failure 404 {object} utils.APIError[any]
// @Failure 500 {object} utils.APIError[any]
func (ic *MonitorController) GetStatPoints(e *echo.Context) error {
	id := e.Param("id")

	sinceStr := e.QueryParam("since")
	if sinceStr == "" {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Missing required 'since' parameter"))
	}
	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid 'since' parameter (must be RFC3339)"))
	}

	untilStr := e.QueryParam("until")
	var until time.Time
	if untilStr == "" {
		until = time.Now().UTC()
	} else {
		until, err = time.Parse(time.RFC3339, untilStr)
		if err != nil {
			return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid 'until' parameter (must be RFC3339)"))
		}
	}

	granularity := e.QueryParam("granularity")
	if granularity == "" {
		granularity = "minute"
	}

	var interval time.Duration
	switch granularity {
	case "minute":
		interval = time.Minute
	case "hour":
		interval = time.Hour
	case "day":
		interval = 24 * time.Hour
	default:
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid 'granularity' parameter (must be minute, hour, or day)"))
	}

	if until.Before(since) {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("'until' must be after 'since'"))
	}

	diff := until.Sub(since)
	estPoints := int(diff/interval) + 1
	if estPoints > 1441 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(fmt.Sprintf("Too many points requested: %d (max 1441)", estPoints)))
	}

	summary, err := ic.monitorService.GetStatPoints(e.Request().Context(), id, since, until, granularity)
	if err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", summary))
}

// @Router /monitors/{id}/stats/uptime [get]
// @Summary Get monitor uptime stats (24h, 30d, 365d)
// @Tags Monitors
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Success 200 {object} utils.ApiResponse[CustomUptimeStatsDto]
// @Failure 400 {object} utils.APIError[any]
// @Failure 404 {object} utils.APIError[any]
// @Failure 500 {object} utils.APIError[any]
func (ic *MonitorController) GetUptimeStats(e *echo.Context) error {
	id := e.Param("id")

	stats, err := ic.monitorService.GetUptimeStats(e.Request().Context(), id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		ic.logger.Errorw("Failed to get uptime stats (short)", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", stats))
}

// @Router		/monitors/batch [get]
// @Summary		Get monitors by IDs
// @Tags			Monitors
// @Produce		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     ids    query     string  true  "Comma-separated list of monitor IDs"
// @Success		200	{object}	utils.ApiResponse[[]Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *MonitorController) FindByIDs(e *echo.Context) error {
	idsStr := e.QueryParam("ids")
	if idsStr == "" {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("ids parameter is required"))
	}

	// Split the comma-separated string into an array
	ids := strings.Split(idsStr, ",")
	if len(ids) == 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("at least one monitor ID is required"))
	}

	// Limit the number of IDs to prevent abuse
	if len(ids) > 100 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("maximum 100 monitor IDs allowed"))
	}

	monitors, err := ic.monitorService.FindByIDs(e.Request().Context(), ids)
	if err != nil {
		ic.logger.Errorw("Failed to fetch monitors", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", monitors))
}

// @Router /monitors/{id}/reset [post]
// @Summary Reset monitor data (heartbeats and stats)
// @Tags Monitors
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Success 200 {object} utils.ApiResponse[any]
// @Failure 400 {object} utils.APIError[any]
// @Failure 404 {object} utils.APIError[any]
// @Failure 500 {object} utils.APIError[any]
func (ic *MonitorController) ResetMonitorData(e *echo.Context) error {
	id := e.Param("id")

	err := ic.monitorService.ResetMonitorData(e.Request().Context(), id)
	if err != nil {
		if err.Error() == "monitor not found" {
			return e.JSON(http.StatusNotFound, utils.NewFailResponse("Monitor not found"))
		}
		ic.logger.Errorw("Failed to reset monitor data", "monitorID", id, "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Monitor data reset successfully", nil))
}

// @Router /monitors/{id}/tls [get]
// @Summary Get monitor TLS certificate information
// @Tags Monitors
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Success 200 {object} utils.ApiResponse[any]
// @Failure 400 {object} utils.APIError[any]
// @Failure 404 {object} utils.APIError[any]
// @Failure 500 {object} utils.APIError[any]
func (ic *MonitorController) GetTLSInfo(e *echo.Context) error {
	id := e.Param("id")

	// First, verify the monitor exists
	_, err := ic.monitorService.FindByID(e.Request().Context(), id)
	if err != nil {
		if err.Error() == "monitor not found" {
			return e.JSON(http.StatusNotFound, utils.NewFailResponse("Monitor not found"))
		}
		ic.logger.Errorw("Failed to fetch monitor", "monitorID", id, "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	// Get TLS info for the monitor
	tlsInfo, err := ic.tlsInfoService.GetTLSInfo(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to get TLS info", "monitorID", id, "error", err) // TODO: fix 500 when no rows in a set
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	// If no TLS info found, return null/empty
	if tlsInfo == nil {
		return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("success", nil))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", tlsInfo))
}
