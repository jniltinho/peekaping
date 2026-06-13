package notification_channel

import (
	"net/http"
	"peekaping/internal/modules/heartbeat"
	"peekaping/internal/modules/monitor"
	"peekaping/internal/modules/shared"
	"peekaping/internal/utils"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type Controller struct {
	service Service
	logger  *zap.SugaredLogger
}

func NewController(
	service Service,
	logger *zap.SugaredLogger,
) *Controller {
	return &Controller{
		service,
		logger,
	}
}

// @Router		/notification-channels [get]
// @Summary		Get notification channels
// @Tags			Notification channels
// @Produce		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     q    query     string  false  "Search query"
// @Param     page query     int     false  "Page number" default(1)
// @Param     limit query    int     false  "Items per page" default(10)
// @Success		200	{object}	utils.ApiResponse[[]Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) FindAll(e *echo.Context) error {
	// Extract query parameters for pagination and search
	page, err := utils.GetQueryInt(e, "page", 0)
	if err != nil || page < 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid page parameter"))
	}

	limit, err := utils.GetQueryInt(e, "limit", 10)
	if err != nil || limit < 1 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid limit parameter"))
	}

	q := e.QueryParam("q")

	response, err := ic.service.FindAll(e.Request().Context(), page, limit, q)
	if err != nil {
		ic.logger.Errorw("Failed to fetch notifications", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", response))
}

// @Router		/notification-channels [post]
// @Summary		Create notification channel
// @Tags			Notification channels
// @Produce		json
// @Accept		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     body body   CreateUpdateDto  true  "Notification object"
// @Success		201	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Create(e *echo.Context) error {
	var notification_channel *CreateUpdateDto
	if err := e.Bind(&notification_channel); err != nil {
		ic.logger.Errorw("Invalid request body", "error", err)
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}

	if err := utils.Validate.Struct(notification_channel); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}

	integration, ok := GetNotificationChannelProvider(notification_channel.Type)
	if !ok {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Unsupported notification type"))
	}
	err := integration.Validate(notification_channel.Config)
	if err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid config: "+err.Error()))
	}

	createdNotification, err := ic.service.Create(e.Request().Context(), notification_channel)
	if err != nil {
		ic.logger.Errorw("Failed to create notification", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusCreated, utils.NewSuccessResponse("Notification created successfully", createdNotification))
}

// @Router		/notification-channels/{id} [get]
// @Summary		Get notification channel by ID
// @Tags			Notification channels
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Notification ID"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) FindByID(e *echo.Context) error {
	id := e.Param("id")

	notification, err := ic.service.FindByID(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to fetch notification", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	if notification == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Notification not found"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", notification))
}

// @Router		/notification-channels/{id} [put]
// @Summary		Update notification channel
// @Tags			Notification channels
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Notification ID"
// @Param       notification body     CreateUpdateDto  true  "Notification object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) UpdateFull(e *echo.Context) error {
	id := e.Param("id")

	var notification CreateUpdateDto
	if err := e.Bind(&notification); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := utils.Validate.Struct(notification); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	updatedNotification, err := ic.service.UpdateFull(e.Request().Context(), id, &notification)
	if err != nil {
		ic.logger.Errorw("Failed to update notification", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("notification updated successfully", updatedNotification))
}

// @Router		/notification-channels/{id} [patch]
// @Summary		Update notification channel
// @Tags			Notification channels
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Notification ID"
// @Param       notification body     PartialUpdateDto  true  "Notification object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) UpdatePartial(e *echo.Context) error {
	id := e.Param("id")

	var notification PartialUpdateDto
	if err := e.Bind(&notification); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// validate
	if err := utils.Validate.Struct(notification); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	updatedNotification, err := ic.service.UpdatePartial(e.Request().Context(), id, &notification)
	if err != nil {
		ic.logger.Errorw("Failed to update notification", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("notification updated successfully", updatedNotification))
}

// @Router		/notification-channels/{id} [delete]
// @Summary		Delete notification channel
// @Tags			Notification channels
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Notification ID"
// @Success		200	{object}	utils.ApiResponse[any]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Delete(e *echo.Context) error {
	id := e.Param("id")

	err := ic.service.Delete(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to delete notification", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Notification deleted successfully", nil))
}

// @Router		/notification-channels/test [post]
// @Summary		Test notification channel
// @Tags			Notification channels
// @Produce		json
// @Accept		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     body body   CreateUpdateDto  true  "Notification object"
// @Success		200	{object}	utils.ApiResponse[any]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Test(e *echo.Context) error {
	var notificationChannel *CreateUpdateDto
	if err := e.Bind(&notificationChannel); err != nil {
		ic.logger.Errorw("Invalid request body", "error", err)
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}

	if err := utils.Validate.Struct(notificationChannel); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}

	integration, ok := GetNotificationChannelProvider(notificationChannel.Type)
	if !ok {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Unsupported notification type"))
	}
	err := integration.Validate(notificationChannel.Config)
	if err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid config: "+err.Error()))
	}

	// Create a test message and monitor for the notification
	testMessage := "This is a test notification from Monitoring"
	testMonitor := &monitor.Model{
		Name: "Test Monitor",
		Type: "http",
	}
	testHeartbeat := &heartbeat.Model{
		Status: shared.MonitorStatusDown,
		Msg:    testMessage,
	}

	// Send the test notification
	err = integration.Send(e.Request().Context(), notificationChannel.Config, testMessage, testMonitor, testHeartbeat)
	if err != nil {
		ic.logger.Errorw("Failed to send test notification", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Failed to send test notification: "+err.Error()))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Test notification sent successfully", nil))
}
