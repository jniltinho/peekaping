package status_page

import (
	"net/http"
	"peekaping/internal/modules/heartbeat"
	"peekaping/internal/modules/monitor"
	"peekaping/internal/utils"
	"time"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type Controller struct {
	service          Service
	monitorService   monitor.Service
	heartbeatService heartbeat.Service
	logger           *zap.SugaredLogger
}

func NewController(service Service, monitorService monitor.Service, heartbeatService heartbeat.Service, logger *zap.SugaredLogger) *Controller {
	return &Controller{
		service:          service,
		monitorService:   monitorService,
		heartbeatService: heartbeatService,
		logger:           logger,
	}
}

// @Router    /status-pages [post]
// @Summary   Create a new status page
// @Tags      Status Pages
// @Accept    json
// @Produce   json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     body body CreateStatusPageDTO true "Status Page object"
// @Success   201  {object} utils.ApiResponse[Model]
// @Failure   400  {object} utils.APIError[any]
// @Failure   500  {object} utils.APIError[any]
func (c *Controller) Create(e *echo.Context) error {
	var dto CreateStatusPageDTO
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := utils.Validate.Struct(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	created, err := c.service.Create(e.Request().Context(), &dto)
	if err != nil {
		// Surface domain uniqueness validation errors as 400
		if domainErr, ok := err.(*DomainAlreadyUsedError); ok {
			return e.JSON(http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":   domainErr.Code,
					"domain": domainErr.Domain,
				},
			})
		}
		c.logger.Errorw("Failed to create status page", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusCreated, utils.NewSuccessResponse("Status page created successfully", created))
}

// @Router    /status-pages/{id} [get]
// @Summary   Get a status page by ID
// @Tags      Status Pages
// @Produce   json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     id   path      string  true  "Status Page ID"
// @Success   200  {object}  utils.ApiResponse[StatusPageWithMonitorsResponseDTO]
// @Failure   404  {object}  utils.APIError[any]
// @Failure   500  {object}  utils.APIError[any]
func (c *Controller) FindByID(e *echo.Context) error {
	id := e.Param("id")
	page, err := c.service.FindByIDWithMonitors(e.Request().Context(), id)
	if err != nil {
		c.logger.Errorw("Failed to get status page by id", "error", err, "id", id)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if page == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Status page not found"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", page))
}

// @Router    /status-pages/slug/{slug} [get]
// @Summary   Get a status page by slug
// @Tags      Status Pages
// @Produce   json
// @Param     slug path      string  true  "Status Page Slug"
// @Success   200  {object}  utils.ApiResponse[Model]
// @Failure   404  {object}  utils.APIError[any]
// @Failure   500  {object}  utils.APIError[any]
func (c *Controller) FindBySlug(e *echo.Context) error {
	slug := e.Param("slug")
	page, err := c.service.FindBySlug(e.Request().Context(), slug)
	if err != nil {
		c.logger.Errorw("Failed to get status page by slug", "error", err, "slug", slug)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if page == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Status page not found"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", page))
}

// @Router    /status-pages/domain/{domain} [get]
// @Summary   Get a status page by domain name
// @Tags      Status Pages
// @Produce   json
// @Param     domain path      string  true  "Domain Name"
// @Success   200  {object}  utils.ApiResponse[Model]
// @Failure   404  {object}  utils.APIError[any]
// @Failure   500  {object}  utils.APIError[any]
func (c *Controller) FindByDomain(e *echo.Context) error {
	domain := e.Param("domain")
	page, err := c.service.FindByDomain(e.Request().Context(), domain)
	if err != nil {
		c.logger.Errorw("Failed to get status page by domain", "error", err, "domain", domain)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if page == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Status page not found"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", page))
}

// @Router    /status-pages [get]
// @Summary   Get all status pages
// @Tags      Status Pages
// @Produce   json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     q    query     string  false  "Search query"
// @Param     page query     int     false  "Page number" default(0)
// @Param     limit query    int     false  "Items per page" default(10)
// @Success   200  {object}  utils.ApiResponse[[]Model]
// @Failure   400  {object}  utils.APIError[any]
// @Failure   500  {object}  utils.APIError[any]
func (c *Controller) FindAll(e *echo.Context) error {
	page, err := utils.GetQueryInt(e, "page", 0)
	if err != nil || page < 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid page parameter"))
	}
	limit, err := utils.GetQueryInt(e, "limit", 10)
	if err != nil || limit < 1 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid limit parameter"))
	}
	q := e.QueryParam("q")

	pages, err := c.service.FindAll(e.Request().Context(), page, limit, q)
	if err != nil {
		c.logger.Errorw("Failed to get all status pages", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", pages))
}

// @Router    /status-pages/{id} [patch]
// @Summary   Update a status page
// @Tags      Status Pages
// @Accept    json
// @Produce   json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     id   path      string  true  "Status Page ID"
// @Param     body body UpdateStatusPageDTO true "Status Page object"
// @Success   200  {object}  utils.ApiResponse[Model]
// @Failure   400  {object}  utils.APIError[any]
// @Failure   404  {object}  utils.APIError[any]
// @Failure   500  {object}  utils.APIError[any]
func (c *Controller) Update(e *echo.Context) error {
	id := e.Param("id")
	var dto UpdateStatusPageDTO
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	updated, err := c.service.Update(e.Request().Context(), id, &dto)
	if err != nil {
		// Surface domain uniqueness validation errors as 400
		if domainErr, ok := err.(*DomainAlreadyUsedError); ok {
			return e.JSON(http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":   domainErr.Code,
					"domain": domainErr.Domain,
				},
			})
		}
		c.logger.Errorw("Failed to update status page", "error", err, "id", id)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("Status page updated successfully", updated))
}

// @Router    /status-pages/{id} [delete]
// @Summary   Delete a status page
// @Tags      Status Pages
// @Produce   json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     id   path      string  true  "Status Page ID"
// @Success   200  {object}  utils.ApiResponse[any]
// @Failure   404  {object}  utils.APIError[any]
// @Failure   500  {object}  utils.APIError[any]
func (c *Controller) Delete(e *echo.Context) error {
	id := e.Param("id")
	err := c.service.Delete(e.Request().Context(), id)
	if err != nil {
		c.logger.Errorw("Failed to delete status page", "error", err, "id", id)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Status page deleted successfully", nil))
}

// @Router    /status-pages/slug/{slug}/monitors [get]
// @Summary   Get monitors for a status page by slug with heartbeats and uptime
// @Tags      Status Pages
// @Produce   json
// @Param     slug path      string  true  "Status Page Slug"
// @Success   200  {object}  utils.ApiResponse[[]MonitorWithHeartbeatsAndUptimeDTO]
// @Failure   404  {object}  utils.APIError[any]
// @Failure   500  {object}  utils.APIError[any]
func (c *Controller) GetMonitorsBySlug(e *echo.Context) error {
	slug := e.Param("slug")

	// First get the status page
	page, err := c.service.FindBySlug(e.Request().Context(), slug)
	if err != nil {
		c.logger.Errorw("Failed to get status page by slug", "error", err, "slug", slug)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if page == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Status page not found"))
	}

	// Get monitors for the status page
	monitors, err := c.service.GetMonitorsForStatusPage(e.Request().Context(), page.ID)
	if err != nil {
		c.logger.Errorw("Failed to get monitors for status page", "error", err, "statusPageID", page.ID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	// Convert monitor_status_page models to monitor models with heartbeats and uptime
	monitorModels := make([]*MonitorWithHeartbeatsAndUptimeDTO, 0, len(monitors))
	for _, msp := range monitors {
		// Get the actual monitor data
		monitorModel, err := c.monitorService.FindByID(e.Request().Context(), msp.MonitorID)
		if err != nil {
			c.logger.Errorw("Failed to get monitor by ID", "error", err, "monitorID", msp.MonitorID)
			continue
		}
		if monitorModel == nil {
			continue
		}

		// Get 100 heartbeats for this monitor
		heartbeats, err := c.heartbeatService.FindByMonitorIDPaginated(e.Request().Context(), msp.MonitorID, 100, 0, nil, true)
		if err != nil {
			c.logger.Errorw("Failed to get heartbeats for monitor", "error", err, "monitorID", msp.MonitorID)
			heartbeats = []*heartbeat.Model{} // Empty slice if error
		}

		// Convert heartbeats to public DTOs
		publicHeartbeats := make([]*PublicHeartbeatDTO, 0, len(heartbeats))
		for _, hb := range heartbeats {
			publicHeartbeat := &PublicHeartbeatDTO{
				ID:      hb.ID,
				Status:  hb.Status,
				Time:    hb.Time,
				EndTime: hb.EndTime,
				Ping:    hb.Ping,
			}
			publicHeartbeats = append(publicHeartbeats, publicHeartbeat)
		}

		// Get 24h uptime for this monitor
		now := time.Now().UTC()
		periods := map[string]time.Duration{
			"24h": 24 * time.Hour,
		}
		uptimeStats, err := c.heartbeatService.FindUptimeStatsByMonitorID(e.Request().Context(), msp.MonitorID, periods, now)
		if err != nil {
			c.logger.Errorw("Failed to get uptime stats for monitor", "error", err, "monitorID", msp.MonitorID)
			return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("failed to get uptime stats for monitor"))
		}

		uptime24h := 0.0
		if uptimeStats != nil {
			if uptime, exists := uptimeStats["24h"]; exists {
				uptime24h = uptime
			}
		}

		publicMonitor := &PublicMonitorDTO{
			ID:     monitorModel.ID,
			Type:   monitorModel.Type,
			Name:   monitorModel.Name,
			Active: monitorModel.Active,
		}

		monitorWithData := &MonitorWithHeartbeatsAndUptimeDTO{
			PublicMonitorDTO: publicMonitor,
			Heartbeats:       publicHeartbeats,
			Uptime24h:        uptime24h,
		}

		monitorModels = append(monitorModels, monitorWithData)
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", monitorModels))
}

// @Router    /status-pages/slug/{slug}/monitors/homepage [get]
// @Summary   Get monitors for a status page by slug for homepage
// @Tags      Status Pages
// @Produce   json
// @Param     slug path      string  true  "Status Page Slug"
// @Success   200  {object}  utils.ApiResponse[[]MonitorWithHeartbeatsAndUptimeDTO]
// @Failure   404  {object}  utils.APIError[any]
// @Failure   500  {object}  utils.APIError[any]
func (c *Controller) GetMonitorsBySlugForHomepage(e *echo.Context) error {
	slug := e.Param("slug")

	// First get the status page
	page, err := c.service.FindBySlug(e.Request().Context(), slug)
	if err != nil {
		c.logger.Errorw("Failed to get status page by slug", "error", err, "slug", slug)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if page == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Status page not found"))
	}

	// Get monitors for the status page
	monitors, err := c.service.GetMonitorsForStatusPage(e.Request().Context(), page.ID)
	if err != nil {
		c.logger.Errorw("Failed to get monitors for status page", "error", err, "statusPageID", page.ID)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	// Convert monitor_status_page models to monitor models with heartbeats and uptime
	monitorModels := make([]*MonitorWithHeartbeatsAndUptimeDTO, 0, len(monitors))
	for _, msp := range monitors {
		// Get the actual monitor data
		monitorModel, err := c.monitorService.FindByID(e.Request().Context(), msp.MonitorID)
		if err != nil {
			c.logger.Errorw("Failed to get monitor by ID", "error", err, "monitorID", msp.MonitorID)
			continue
		}
		if monitorModel == nil {
			continue
		}

		// Get 100 heartbeats for this monitor
		heartbeats, err := c.heartbeatService.FindByMonitorIDPaginated(e.Request().Context(), msp.MonitorID, 1, 0, nil, true)
		if err != nil {
			c.logger.Errorw("Failed to get heartbeats for monitor", "error", err, "monitorID", msp.MonitorID)
			heartbeats = []*heartbeat.Model{} // Empty slice if error
		}

		// Convert heartbeats to public DTOs
		publicHeartbeats := make([]*PublicHeartbeatDTO, 0, len(heartbeats))
		for _, hb := range heartbeats {
			publicHeartbeat := &PublicHeartbeatDTO{
				ID:      hb.ID,
				Status:  hb.Status,
				Time:    hb.Time,
				EndTime: hb.EndTime,
				Ping:    hb.Ping,
			}
			publicHeartbeats = append(publicHeartbeats, publicHeartbeat)
		}

		// Get 24h uptime for this monitor
		now := time.Now().UTC()
		periods := map[string]time.Duration{
			"24h": 24 * time.Hour,
		}
		uptimeStats, err := c.heartbeatService.FindUptimeStatsByMonitorID(e.Request().Context(), msp.MonitorID, periods, now)
		if err != nil {
			c.logger.Errorw("Failed to get uptime stats for monitor", "error", err, "monitorID", msp.MonitorID)
			return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("failed to get uptime stats for monitor"))
		}

		uptime24h := 0.0
		if uptimeStats != nil {
			if uptime, exists := uptimeStats["24h"]; exists {
				uptime24h = uptime
			}
		}

		publicMonitor := &PublicMonitorDTO{
			ID:     monitorModel.ID,
			Type:   monitorModel.Type,
			Name:   monitorModel.Name,
			Active: monitorModel.Active,
		}

		monitorWithData := &MonitorWithHeartbeatsAndUptimeDTO{
			PublicMonitorDTO: publicMonitor,
			Heartbeats:       publicHeartbeats,
			Uptime24h:        uptime24h,
		}

		monitorModels = append(monitorModels, monitorWithData)
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", monitorModels))
}
