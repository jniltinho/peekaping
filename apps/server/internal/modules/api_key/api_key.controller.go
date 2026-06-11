package api_key

import (
	"net/http"
	"peekaping/internal/utils"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/go-playground/validator/v10"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{
		service: service,
	}
}

// CreateAPIKey creates a new API key
// @Summary Create API key
// @Description Create a new API key
// @Tags api-keys
// @Accept json
// @Produce json
// @Security JwtAuth
// @Param request body CreateAPIKeyDto true "API key creation data"
// @Success 201 {object} utils.ApiResponse[APIKeyWithTokenResponse]
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api-keys [post]
func (c *Controller) CreateAPIKey(e *echo.Context) error {
	// MARK: CreateAPIKey

	var req CreateAPIKeyDto
	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// Validate the request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// Validate expiration date
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("expiration date cannot be in the past"))
	}

	// Validate max usage count
	if req.MaxUsageCount != nil && *req.MaxUsageCount <= 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("max usage count must be greater than 0"))
	}

	// Convert DTO to service request
	serviceReq := &CreateRequest{
		Name:          req.Name,
		ExpiresAt:     req.ExpiresAt,
		MaxUsageCount: req.MaxUsageCount,
	}

	apiKey, err := c.service.Create(e.Request().Context(), serviceReq)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse(err.Error()))
	}

	return e.JSON(http.StatusCreated, utils.NewSuccessResponse("API key created successfully", apiKey.ToAPIKeyWithTokenResponse()))
}

// GetAPIKeys gets all API keys
// @Summary Get API keys
// @Description Get all API keys
// @Tags api-keys
// @Produce json
// @Security JwtAuth
// @Success 200 {object} utils.ApiResponse[[]APIKeyResponse]
// @Failure 500 {object} utils.APIError
// @Router /api-keys [get]
func (c *Controller) GetAPIKeys(e *echo.Context) error {
	// MARK: GetAPIKeys

	apiKeys, err := c.service.FindAll(e.Request().Context())
	if err != nil {
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse(err.Error()))
	}

	responses := make([]*APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		responses[i] = apiKey.ToAPIKeyResponse()
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("API keys retrieved successfully", responses))
}

// GetAPIKey gets a specific API key by ID
// @Summary Get API key
// @Description Get a specific API key by ID
// @Tags api-keys
// @Produce json
// @Security JwtAuth
// @Param id path string true "API key ID"
// @Success 200 {object} utils.ApiResponse[APIKeyResponse]
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api-keys/{id} [get]
func (c *Controller) GetAPIKey(e *echo.Context) error {
	// MARK: GetAPIKey

	id := e.Param("id")
	apiKey, err := c.service.FindByID(e.Request().Context(), id)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse(err.Error()))
	}
	if apiKey == nil {
		return e.JSON(http.StatusNotFound, utils.NewFailResponse("API key not found"))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("API key retrieved successfully", apiKey.ToAPIKeyResponse()))
}

// UpdateAPIKey updates an API key
// @Summary Update API key
// @Description Update an API key
// @Tags api-keys
// @Accept json
// @Produce json
// @Security JwtAuth
// @Param id path string true "API key ID"
// @Param request body UpdateAPIKeyDto true "API key update data"
// @Success 200 {object} utils.ApiResponse[APIKeyResponse]
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api-keys/{id} [put]
func (c *Controller) UpdateAPIKey(e *echo.Context) error {
	// MARK: UpdateAPIKey

	id := e.Param("id")
	var req UpdateAPIKeyDto
	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// Validate the request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// Validate expiration date if provided
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("expiration date cannot be in the past"))
	}

	// Validate max usage count if provided
	if req.MaxUsageCount != nil && *req.MaxUsageCount <= 0 {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("max usage count must be greater than 0"))
	}

	// Convert DTO to service request
	serviceReq := &UpdateRequest{
		Name:          req.Name,
		ExpiresAt:     req.ExpiresAt,
		MaxUsageCount: req.MaxUsageCount,
	}

	apiKey, err := c.service.Update(e.Request().Context(), id, serviceReq)
	if err != nil {
		if err.Error() == "API key not found" {
			return e.JSON(http.StatusNotFound, utils.NewFailResponse("API key not found"))
		}
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse(err.Error()))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("API key updated successfully", apiKey.ToAPIKeyResponse()))
}

// DeleteAPIKey deletes an API key
// @Summary Delete API key
// @Description Delete an API key
// @Tags api-keys
// @Security JwtAuth
// @Param id path string true "API key ID"
// @Success 204 "No Content"
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api-keys/{id} [delete]
func (c *Controller) DeleteAPIKey(e *echo.Context) error {
	// MARK: DeleteAPIKey

	id := e.Param("id")
	err := c.service.Delete(e.Request().Context(), id)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse(err.Error()))
	}

	e.Response().WriteHeader(http.StatusNoContent)
	return nil
}

// GetAPIKeyConfig gets API key configuration
// @Summary Get API key configuration
// @Description Get API key configuration including prefix
// @Tags api-keys
// @Produce json
// @Success 200 {object} utils.ApiResponse[APIKeyConfigResponse]
// @Router /api-keys/config [get]
func (c *Controller) GetAPIKeyConfig(e *echo.Context) error {
	// MARK: GetAPIKeyConfig

	config := &APIKeyConfigResponse{
		Prefix: ApiKeyPrefix,
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse("API key configuration retrieved successfully", config))
}
