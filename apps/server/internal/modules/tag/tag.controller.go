package tag

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
	return &Controller{
		service,
		logger,
	}
}

// @Router		/tags [get]
// @Summary		Get tags
// @Tags			Tags
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

	response, err := c.service.FindAll(e.Request().Context(), page, limit, q)
	if err != nil {
		c.logger.Errorw("Failed to fetch tags", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", response))
}

// @Router		/tags [post]
// @Summary		Create tag
// @Tags			Tags
// @Produce		json
// @Accept		json
// @Security  JwtAuth
// @Security  ApiKeyAuth
// @Param     body body   CreateUpdateDto  true  "Tag object"
// @Success		201	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) Create(e *echo.Context) error {
	var tag *CreateUpdateDto
	if err := e.Bind(&tag); err != nil {
		c.logger.Errorw("Invalid request body", "error", err)
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}

	if err := utils.Validate.Struct(tag); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	createdTag, err := c.service.Create(e.Request().Context(), tag)
	if err != nil {
		c.logger.Errorw("Failed to create tag", "error", err)
		if err.Error() == "tag with this name already exists" {
			return e.JSON(http.StatusConflict, utils.NewFailResponse("Tag with this name already exists"))
		}
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusCreated, utils.NewSuccessResponse("Tag created successfully", createdTag))
}

// @Router		/tags/{id} [get]
// @Summary		Get tag by ID
// @Tags			Tags
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Tag ID"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) FindByID(e *echo.Context) error {
	id := e.Param("id")

	tag, err := c.service.FindByID(e.Request().Context(), id)
	if err != nil {
		c.logger.Errorw("Failed to fetch tag", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	if tag == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Tag not found"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", tag))
}

// @Router		/tags/{id} [put]
// @Summary		Update tag
// @Tags			Tags
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Tag ID"
// @Param       tag body     CreateUpdateDto  true  "Tag object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) UpdateFull(e *echo.Context) error {
	id := e.Param("id")

	var tag CreateUpdateDto
	if err := e.Bind(&tag); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := utils.Validate.Struct(tag); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	updatedTag, err := c.service.UpdateFull(e.Request().Context(), id, &tag)
	if err != nil {
		c.logger.Errorw("Failed to update tag", "error", err)
		if err.Error() == "tag with this name already exists" {
			return e.JSON(http.StatusConflict, utils.NewFailResponse("Tag with this name already exists"))
		}
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("Tag updated successfully", updatedTag))
}

// @Router		/tags/{id} [patch]
// @Summary		Update tag
// @Tags			Tags
// @Produce		json
// @Accept		json
// @Security BearerAuth
// @Param       id   path      string  true  "Tag ID"
// @Param       tag body     PartialUpdateDto  true  "Tag object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) UpdatePartial(e *echo.Context) error {
	id := e.Param("id")

	var tag PartialUpdateDto
	if err := e.Bind(&tag); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := utils.Validate.Struct(tag); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	updatedTag, err := c.service.UpdatePartial(e.Request().Context(), id, &tag)
	if err != nil {
		c.logger.Errorw("Failed to update tag", "error", err)
		if err.Error() == "tag with this name already exists" {
			return e.JSON(http.StatusConflict, utils.NewFailResponse("Tag with this name already exists"))
		}
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("Tag updated successfully", updatedTag))
}

// @Router		/tags/{id} [delete]
// @Summary		Delete tag
// @Tags			Tags
// @Produce		json
// @Security BearerAuth
// @Param       id   path      string  true  "Tag ID"
// @Success		200	{object}	utils.ApiResponse[any]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) Delete(e *echo.Context) error {
	id := e.Param("id")

	err := c.service.Delete(e.Request().Context(), id)
	if err != nil {
		c.logger.Errorw("Failed to delete tag", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Tag deleted successfully", nil))
}
