package maintenance

import (
	"fmt"
	"net/http"
	"peekaping/internal/utils"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/bson"
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

// @Router		/maintenances [get]
// @Summary		Get maintenances
// @Tags			Maintenances
// @Produce		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     q    query     string  false  "Search query"
// @Param     strategy query string  false  "Filter by strategy"
// @Param     page query     int     false  "Page number" default(1)
// @Param     limit query    int     false  "Items per page" default(10)
// @Success		200	{object}	utils.ApiResponse[[]Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) FindAll(e *echo.Context) error {
	page, err := utils.GetQueryInt(e, "page", 0)
	if err != nil || page < 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid page parameter"))
	}

	limit, err := utils.GetQueryInt(e, "limit", 10)
	if err != nil || limit < 1 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid limit parameter"))
	}

	q := e.QueryParam("q")
	strategy := e.QueryParam("strategy")

	entities, err := ic.service.FindAll(e.Request().Context(), page, limit, q, strategy)
	if err != nil {
		ic.logger.Errorw("Failed to fetch maintenances", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", entities))
}

// @Router		/maintenances [post]
// @Summary		Create maintenance
// @Tags			Maintenances
// @Produce		json
// @Accept		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     body body   CreateUpdateDto  true  "Maintenance object"
// @Success		201	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Create(ctx *gin.Context) {
	var entity *CreateUpdateDto
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	if err := utils.Validate.Struct(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	created, err := ic.service.Create(ctx, entity)
	if err != nil {
		ic.logger.Errorw("Failed to create maintenance", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	ctx.JSON(http.StatusCreated, utils.NewSuccessResponse("Maintenance created successfully", created))
}

// @Router		/maintenances/{id} [get]
// @Summary		Get maintenance by ID
// @Tags			Maintenances
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Maintenance ID"
// @Success		200	{object}	utils.ApiResponse[MaintenanceResponseDto]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) FindByID(ctx *gin.Context) {
	id := ctx.Param("id")

	entity, err := ic.service.FindByID(ctx, id)
	if err != nil {
		ic.logger.Errorw("Failed to fetch maintenance", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	if entity == nil {
		ctx.JSON(http.StatusNotFound, utils.NewFailResponse("Maintenance not found"))
		return
	}

	// Get monitor IDs
	monitorIds, err := ic.service.GetMonitors(ctx, id)
	if err != nil {
		ic.logger.Errorw("Failed to fetch monitor IDs", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	response := &MaintenanceResponseDto{
		ID:            entity.ID,
		Title:         entity.Title,
		Description:   entity.Description,
		Active:        entity.Active,
		Strategy:      entity.Strategy,
		StartDateTime: entity.StartDateTime,
		EndDateTime:   entity.EndDateTime,
		StartTime:     entity.StartTime,
		EndTime:       entity.EndTime,
		Weekdays:      entity.Weekdays,
		DaysOfMonth:   entity.DaysOfMonth,
		IntervalDay:   entity.IntervalDay,
		Cron:          entity.Cron,
		Timezone:      entity.Timezone,
		Duration:      entity.Duration,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
		MonitorIds:    monitorIds,
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("success", response))
}

// @Router		/maintenances/{id} [put]
// @Summary		Update maintenance
// @Tags			Maintenances
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Maintenance ID"
// @Param       body body     CreateUpdateDto  true  "Maintenance object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) UpdateFull(ctx *gin.Context) {
	id := ctx.Param("id")

	var entity CreateUpdateDto
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	if err := utils.Validate.Struct(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	updated, err := ic.service.UpdateFull(ctx, id, &entity)
	if err != nil {
		ic.logger.Errorw("Failed to update maintenance", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("maintenance updated successfully", updated))
}

// @Router		/maintenances/{id} [patch]
// @Summary		Update maintenance
// @Tags			Maintenances
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Maintenance ID"
// @Param       body body     PartialUpdateDto  true  "Maintenance object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) UpdatePartial(ctx *gin.Context) {
	id := ctx.Param("id")

	var entity PartialUpdateDto
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	updated, err := ic.service.UpdatePartial(ctx, id, &entity)
	if err != nil {
		ic.logger.Errorw("Failed to update maintenance", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("maintenance updated successfully", updated))
}

// @Router		/maintenances/{id} [delete]
// @Summary		Delete maintenance
// @Tags			Maintenances
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Maintenance ID"
// @Success		200	{object}	utils.ApiResponse[any]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	err := ic.service.Delete(ctx, id)
	if err != nil {
		ic.logger.Errorw("Failed to delete maintenance", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Maintenance deleted successfully", nil))
}

// @Router		/maintenances/{id}/pause [patch]
// @Summary		Pause maintenance
// @Tags			Maintenances
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Maintenance ID"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Pause(ctx *gin.Context) {
	fmt.Println("Pausing maintenance")
	id := ctx.Param("id")
	updated, err := ic.service.SetActive(ctx, id, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Failed to pause maintenance"))
		return
	}
	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("Paused", updated))
}

// @Router		/maintenances/{id}/resume [patch]
// @Summary		Resume maintenance
// @Tags			Maintenances
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Maintenance ID"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Resume(ctx *gin.Context) {
	id := ctx.Param("id")
	updated, err := ic.service.SetActive(ctx, id, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Failed to resume maintenance"))
		return
	}
	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("Resumed", updated))
}
