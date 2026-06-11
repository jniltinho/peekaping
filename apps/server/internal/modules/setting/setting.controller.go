package setting

import (
	"fmt"
	"net/http"
	"peekaping/internal/utils"

	"regexp"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

var screamingSnakeCase = regexp.MustCompile(`^[A-Z0-9_]+$`)

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

// @Router	/settings/key/{key} [get]
// @Summary	Get setting by key
// @Tags		Settings
// @Produce	json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param	key	path	string	true	"Setting Key"
// @Success	200	{object}	utils.ApiResponse[Model]
// @Failure	404	{object}	utils.APIError[any]
// @Failure	500	{object}	utils.APIError[any]
func (ic *Controller) GetByKey(e *echo.Context) error {
	key := e.Param("key")
	if !screamingSnakeCase.MatchString(key) {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Key must be SCREAMING_SNAKE_CASE (A-Z, 0-9, _)."))
	}
	entity, err := ic.service.GetByKey(e.Request().Context(), key)
	if err != nil {
		ic.logger.Errorw("Failed to fetch setting by key", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	if entity == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("Setting not found"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("success", entity))
}

// @Router	/settings/key/{key} [put]
// @Summary	Set setting by key
// @Tags		Settings
// @Produce	json
// @Accept	json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param	key	path	string	true	"Setting Key"
// @Param	body	body	CreateUpdateDto	true	"Setting object"
// @Success	200	{object}	utils.ApiResponse[Model]
// @Failure	400	{object}	utils.APIError[any]
// @Failure	500	{object}	utils.APIError[any]
func (ic *Controller) SetByKey(e *echo.Context) error {
	fmt.Println("SetByKey")
	key := e.Param("key")
	if !screamingSnakeCase.MatchString(key) {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Key must be SCREAMING_SNAKE_CASE (A-Z, 0-9, _)."))
	}
	var entity CreateUpdateDto
	if err := e.Bind(&entity); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}
	if err := utils.Validate.Struct(entity); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}
	updated, err := ic.service.SetByKey(e.Request().Context(), key, &entity)
	if err != nil {
		ic.logger.Errorw("Failed to set setting by key", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("setting set successfully", updated))
}

// @Router	/settings/key/{key} [delete]
// @Summary	Delete setting by key
// @Tags		Settings
// @Produce	json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param	key	path	string	true	"Setting Key"
// @Success	200	{object}	utils.ApiResponse[any]
// @Failure	404	{object}	utils.APIError[any]
// @Failure	500	{object}	utils.APIError[any]
func (ic *Controller) DeleteByKey(e *echo.Context) error {
	key := e.Param("key")
	if !screamingSnakeCase.MatchString(key) {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Key must be SCREAMING_SNAKE_CASE (A-Z, 0-9, _)."))
	}
	err := ic.service.DeleteByKey(e.Request().Context(), key)
	if err != nil {
		ic.logger.Errorw("Failed to delete setting by key", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Setting deleted successfully", nil))
}
