package utils

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetQueryInt extracts an integer query parameter from the context.
// If the parameter is not provided or invalid, it returns the default value.
// Ported from Gin to Echo (QueryParam instead of Query).
func GetQueryInt(c echo.Context, key string, defaultValue int) (int, error) {
	valueStr := c.QueryParam(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// GetQueryBool extracts a boolean query parameter from the context.
// If the parameter is not provided, it returns nil.
// Ported from Gin to Echo.
func GetQueryBool(c echo.Context, key string) (*bool, error) {
	valueStr := c.QueryParam(key)
	if valueStr == "" {
		return nil, nil
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
