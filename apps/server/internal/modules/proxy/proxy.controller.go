package proxy

import (
	"net/http"
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
	// Register custom struct-level validation if needed
	// validate.RegisterStructValidation(CreateUpdateDtoStructLevelValidation, CreateUpdateDto{})
	return &Controller{
		service,
		logger,
	}
}

// @Router		/proxies [get]
// @Summary		Get proxies
// @Tags			Proxies
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
	page, err := utils.GetQueryInt(e, "page", 0)
	if err != nil || page < 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid page parameter"))
	}

	limit, err := utils.GetQueryInt(e, "limit", 10)
	if err != nil || limit < 1 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid limit parameter"))
	}

	q := e.QueryParam("q")

	entities, err := ic.service.FindAll(e.Request().Context(), page, limit, q)
	if err != nil {
		ic.logger.Errorw("Failed to fetch proxies", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", entities))
}

// @Router		/proxies [post]
// @Summary		Create proxy
// @Tags			Proxies
// @Produce		json
// @Accept		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     body body   CreateUpdateDto  true  "Proxy object"
// @Success		201	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Create(e *echo.Context) error {
	var entity *CreateUpdateDto
	if err := e.Bind(&entity); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := utils.Validate.Struct(entity); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	created, err := ic.service.Create(e.Request().Context(), entity)
	if err != nil {
		ic.logger.Errorw("Failed to create proxy", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusCreated, utils.NewSuccessResponse("Proxy created successfully", created))
}

// @Router		/proxies/{id} [get]
// @Summary		Get proxy by ID
// @Tags			Proxies
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Proxy ID"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) FindByID(e *echo.Context) error {
	id := e.Param("id")

	entity, err := ic.service.FindByID(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to fetch proxy", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	if entity == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Proxy not found"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", entity))
}

// @Router		/proxies/{id} [put]
// @Summary		Update proxy
// @Tags			Proxies
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Proxy ID"
// @Param       body body     CreateUpdateDto  true  "Proxy object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) UpdateFull(e *echo.Context) error {
	id := e.Param("id")

	var entity CreateUpdateDto
	if err := e.Bind(&entity); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}

	if err := utils.Validate.Struct(entity); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	updated, err := ic.service.UpdateFull(e.Request().Context(), id, &entity)
	if err != nil {
		ic.logger.Errorw("Failed to update proxy", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("proxy updated successfully", updated))
}

// @Router		/proxies/{id} [patch]
// @Summary		Update proxy
// @Tags			Proxies
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Proxy ID"
// @Param       body body     PartialUpdateDto  true  "Proxy object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) UpdatePartial(e *echo.Context) error {
	id := e.Param("id")

	var entity PartialUpdateDto
	if err := e.Bind(&entity); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}

	updated, err := ic.service.UpdatePartial(e.Request().Context(), id, &entity)
	if err != nil {
		ic.logger.Errorw("Failed to update proxy", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("proxy updated successfully", updated))
}

// @Router		/proxies/{id} [delete]
// @Summary		Delete proxy
// @Tags			Proxies
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Proxy ID"
// @Success		200	{object}	utils.ApiResponse[any]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (ic *Controller) Delete(e *echo.Context) error {
	id := e.Param("id")

	err := ic.service.Delete(e.Request().Context(), id)
	if err != nil {
		ic.logger.Errorw("Failed to delete proxy", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Proxy deleted successfully", nil))
}
