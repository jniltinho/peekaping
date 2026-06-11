package bruteforce

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

// statusRecorder wraps ResponseWriter to capture the final status code.
// Needed because in Echo v5, c.Response() returns http.ResponseWriter directly
// (no .Status field like in v4).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// Service interface for bruteforce protection
type Service interface {
	// IsLocked returns current lock (if any).
	IsLocked(ctx context.Context, key string) (bool, time.Time, error)
	// OnFailure atomically updates counters and may set a lock.
	// Returns (locked, until, err).
	OnFailure(ctx context.Context, key string, now time.Time, window time.Duration, max int, lockout time.Duration) (bool, time.Time, error)
	// Reset clears all state for the key (on successful auth).
	Reset(ctx context.Context, key string) error
}

type KeyExtractor func(*echo.Context) (string, error)

type Config struct {
	MaxAttempts int
	Window      time.Duration
	Lockout     time.Duration
	// Which HTTP statuses of the wrapped handler mean "authentication failed"
	FailureStatuses []int
	// Optional custom blocked response (otherwise 429 with Retry-After)
	OnBlocked func(c *echo.Context, retryAfter time.Duration)
}

type Guard struct {
	cfg          Config
	service      Service
	keyExtractor KeyExtractor
	logger       *zap.SugaredLogger
}

func New(cfg Config, service Service, ke KeyExtractor, logger *zap.SugaredLogger) *Guard {
	// sensible defaults
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Lockout <= 0 {
		cfg.Lockout = 1 * time.Minute
	}
	if cfg.FailureStatuses == nil {
		cfg.FailureStatuses = []int{401, 403}
	}
	return &Guard{
		cfg:          cfg,
		service:      service,
		keyExtractor: ke,
		logger:       logger,
	}
}

func (g *Guard) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			key, err := g.keyExtractor(c)
			if err != nil || key == "" {
				// If we cannot extract key, we fallback to IP only.
				key = c.RealIP()
			}

			ctx := c.Request().Context()

			locked, until, err := g.service.IsLocked(ctx, key)
			if err != nil {
				// Fail safe: pass, log/monitor
				g.logger.Errorw("failed to check if locked", "error", err)
				return next(c)
			}
			if locked {
				// Fix race condition: recalculate retry time to avoid negative values
				retryAfter := time.Until(until)
				if retryAfter <= 0 {
					// Lock expired between check and now, allow request to proceed
					g.logger.Debugw("lock expired, allowing request", "key", key)
					return next(c)
				}
				g.block(c, retryAfter)
				return nil
			}

			// Wrap the response writer so we can capture the status written by the handler
			// (required for Echo v5 where c.Response() returns plain http.ResponseWriter)
			rec := &statusRecorder{ResponseWriter: c.Response(), status: 0}
			c.SetResponse(rec)

			err = next(c)

			// After handler runs, decide success/failure by status
			status := rec.status
			if status == 0 {
				status = http.StatusOK
			}
			if g.isFailure(status) {
				now := time.Now()
				// OnFailure atomically handles counting and locking
				locked, until, err := g.service.OnFailure(ctx, key, now, g.cfg.Window, g.cfg.MaxAttempts, g.cfg.Lockout)
				if err != nil {
					// Log error but don't fail the request - the auth failure should still be returned
					g.logger.Errorw("failed to record failure", "key", key, "error", err)
				} else if locked {
					g.logger.Infow("account locked due to too many failures", "key", key, "until", until)
				}
				return nil // do not swallow original error if any; but for now follow original "return after record"
			}

			// Only reset on explicit success statuses (200-299), not on other statuses like 400, 500
			if status >= 200 && status < 300 {
				if resetErr := g.service.Reset(ctx, key); resetErr != nil {
					g.logger.Errorw("failed to reset bruteforce state", "key", key, "error", resetErr)
				}
			}
			return err
		}
	}
}

func (g *Guard) isFailure(status int) bool {
	for _, s := range g.cfg.FailureStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func (g *Guard) block(c *echo.Context, retryAfter time.Duration) {
	if g.cfg.OnBlocked != nil {
		g.cfg.OnBlocked(c, retryAfter)
		return
	}

	// Ensure retry time is not negative
	if retryAfter < 0 {
		retryAfter = 0
	}

	c.Response().Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
	_ = c.JSON(http.StatusTooManyRequests, map[string]any{
		"success":     false,
		"message":     "too many attempts, try later",
		"retry_after": int(retryAfter.Seconds()),
	})
}

// KeyByIPAndBodyField makes a key "<ip>:<lower(username)>"
// It safely reads the field from JSON body without consuming it by preserving the original body.
func KeyByIPAndBodyField(field string) KeyExtractor {
	return func(c *echo.Context) (string, error) {
		ip := c.RealIP()

		// Only process JSON requests
		contentType := c.Request().Header.Get("Content-Type")
		if contentType == "application/json" || strings.Contains(contentType, "application/json") {
			// Read body safely without consuming it
			if c.Request().Body != nil {
				bodyBytes, err := io.ReadAll(c.Request().Body)
				if err != nil {
					// On error, fallback to IP only
					return ip, nil
				}

				// Restore the body for subsequent handlers
				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Try to parse JSON and extract field
				var m map[string]any
				if err := json.Unmarshal(bodyBytes, &m); err == nil {
					if v, ok := m[field]; ok {
						if s, ok := v.(string); ok && s != "" {
							return fmt.Sprintf("%s:%s", ip, strings.ToLower(s)), nil
						}
					}
				}
			}
		}

		// For form requests, try FormValue (Echo equivalent)
		if contentType == "application/x-www-form-urlencoded" {
			if v := c.FormValue(field); v != "" {
				return fmt.Sprintf("%s:%s", ip, strings.ToLower(v)), nil
			}
		}

		// Fallback to IP only
		return ip, nil
	}
}
