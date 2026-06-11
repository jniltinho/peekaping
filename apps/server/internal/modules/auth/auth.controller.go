package auth

import (
	"errors"
	"fmt"
	"net/http"
	"peekaping/internal/utils"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/go-playground/validator/v10"
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
		service: service,
		logger:  logger,
	}
}

// validateWithDetails provides detailed error messages for validation failures
func (c *Controller) validateWithDetails(dto interface{}) error {
	if err := utils.Validate.Struct(dto); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string

			for _, fieldError := range validationErrors {
				field := fieldError.Field()
				tag := fieldError.Tag()

				switch tag {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("%s is required", field))
				case "email":
					errorMessages = append(errorMessages, "Please provide a valid email address")
				case "password":
					errorMessages = append(errorMessages, "Password must be at least 8 characters long and contain uppercase, lowercase, number, and special character")
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("%s validation failed", field))
				}
			}

			return errors.New(strings.Join(errorMessages, "; "))
		}
		return err
	}
	return nil
}

// @Router		/auth/register [post]
// @Summary		Register new admin
// @Tags			Auth
// @Produce		json
// @Accept		json
// @Param       body body     RegisterDto  true  "Registration data"
// @Success		201	{object}	utils.ApiResponse[LoginResponse]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) Register(e *echo.Context) error {
	var dto RegisterDto
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	// validate with detailed error messages
	if err := c.validateWithDetails(dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	response, err := c.service.Register(e.Request().Context(), dto)
	if err != nil {
		c.logger.Errorw("Failed to register admin", "error", err)
		if err.Error() == "admin already exists" {
			return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		}
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse(err.Error()))
	}

	return e.JSON(http.StatusCreated, utils.NewSuccessResponse("User registered successfully", response))
}

// @Router		/auth/login [post]
// @Summary		Login admin
// @Tags			Auth
// @Produce		json
// @Accept		json
// @Param       body body     LoginDto  true  "Login data"
// @Success		200	{object}	utils.ApiResponse[LoginResponse]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		401	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) Login(e *echo.Context) error {
	var dto LoginDto
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := c.validateWithDetails(dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	response, err := c.service.Login(e.Request().Context(), dto)
	if err != nil {
		c.logger.Errorw("Failed to login admin", "error", err)
		return e.JSON(http.StatusUnauthorized, utils.NewFailResponse(err.Error()))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("Login successful", response))
}

// @Router		/auth/refresh [post]
// @Summary		Refresh access token
// @Tags			Auth
// @Produce		json
// @Accept		json
// @Param       body body     RefreshTokenDto  true  "Refresh token data"
// @Success		200	{object}	utils.ApiResponse[LoginResponse]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		401	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) RefreshToken(e *echo.Context) error {
	var dto RefreshTokenDto
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := c.validateWithDetails(dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	response, err := c.service.RefreshToken(e.Request().Context(), dto.RefreshToken)
	if err != nil {
		c.logger.Errorw("Failed to refresh token", "error", err)
		return e.JSON(http.StatusUnauthorized, utils.NewFailResponse(err.Error()))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse("Token refreshed successfully", response))
}

// @Router	/auth/password [put]
// @Summary	Update user password
// @Tags		Auth
// @Produce	json
// @Accept	json
// @Security JwtAuth
// @Param	body body     UpdatePasswordDto  true  "Password update data"
// @Success	200	{object}	utils.ApiResponse[any]
// @Failure	400	{object}	utils.APIError[any]
// @Failure	401	{object}	utils.APIError[any]
// @Failure	500	{object}	utils.APIError[any]
func (c *Controller) UpdatePassword(e *echo.Context) error {
	userId, exists := e.Get("userId")
	if !exists {
		return e.JSON(http.StatusUnauthorized, utils.NewFailResponse("Unauthorized"))
	}

	var dto UpdatePasswordDto
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := c.validateWithDetails(dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	err := c.service.UpdatePassword(e.Request().Context(), userId.(string), dto)
	if err != nil {
		if err.Error() == "current password is incorrect" {
			return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		}
		c.logger.Errorw("Failed to update password", "error", err)
		return e.JSON(http.StatusInternalServerError, utils.NewFailResponse(err.Error()))
	}

	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Password updated successfully", nil))
}

// @Router	/auth/2fa/setup [post]
// @Summary	Enable 2FA (TOTP) for user
// @Tags		Auth
// @Produce	json
// @Accept	json
// @Security JwtAuth
// @Param	body body     TwoFASetupRequestDto  true  "2FA setup request"
// @Success	200 {object} TwoFASetupResponseDto
// @Failure	400 {object} utils.APIError[any]
// @Failure	500 {object} utils.APIError[any]
func (c *Controller) SetupTwoFA(e *echo.Context) error {
	userId, _ := e.Get("userId")

	var dto TwoFASetupRequestDto
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := c.validateWithDetails(dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	secret, provisioningURI, err := c.service.SetupTwoFA(e.Request().Context(), userId.(string), dto.Password)
	if err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}
	return e.JSON(http.StatusOK, TwoFASetupResponseDto{
		Secret:          secret,
		ProvisioningURI: provisioningURI,
	})
}

// @Router	/auth/2fa/verify [post]
// @Summary	Verify 2FA (TOTP) code for user
// @Tags		Auth
// @Produce	json
// @Accept	json
// @Security JwtAuth
// @Param	body body     TwoFAVerifyRequestDto  true  "2FA verify request"
// @Success	200 {object} TwoFAVerifyResponseDto
// @Failure	400 {object} TwoFAVerifyResponseDto
// @Failure	500 {object} utils.APIError[any]
func (c *Controller) VerifyTwoFA(e *echo.Context) error {
	userId, _ := e.Get("userId")

	var dto TwoFAVerifyRequestDto
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	if err := c.validateWithDetails(dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	success, err := c.service.VerifyTwoFA(e.Request().Context(), userId.(string), dto.Code)
	if err != nil {
		return e.JSON(http.StatusBadRequest, TwoFAVerifyResponseDto{Success: false, Message: err.Error()})
	}

	return e.JSON(http.StatusOK, TwoFAVerifyResponseDto{Success: success, Message: "2FA verification successful"})
}

// @Router	/auth/2fa/disable [post]
// @Summary	Disable 2FA (TOTP) for user
// @Tags		Auth
// @Produce	json
// @Accept	json
// @Security JwtAuth
// @Param	body body     TwoFADisableRequestDto  true  "2FA disable request"
// @Success	200 {object} utils.ApiResponse[any]
// @Failure	400 {object} utils.APIError[any]
// @Failure	500 {object} utils.APIError[any]
func (c *Controller) DisableTwoFA(e *echo.Context) error {
	userId, _ := e.Get("userId")

	var dto TwoFADisableRequestDto
	if err := e.Bind(&dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid request body"))
	}

	if err := c.validateWithDetails(dto); err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}

	err := c.service.DisableTwoFA(e.Request().Context(), userId.(string), dto.Password)
	if err != nil {
		return e.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
	}
	return e.JSON(http.StatusOK, utils.NewSuccessResponse[any]("2FA disabled successfully", nil))
}
