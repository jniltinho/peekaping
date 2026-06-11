package bruteforce

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"peekaping/internal/config"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockService implements the Service interface for Guard tests
// Only methods used by Guard are implemented
type MockService struct {
	mock.Mock
}

func (m *MockService) IsLocked(ctx context.Context, key string) (bool, time.Time, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Get(1).(time.Time), args.Error(2)
}
func (m *MockService) OnFailure(ctx context.Context, key string, now time.Time, window time.Duration, max int, lockout time.Duration) (bool, time.Time, error) {
	args := m.Called(ctx, key, now, window, max, lockout)
	return args.Bool(0), args.Get(1).(time.Time), args.Error(2)
}
func (m *MockService) Reset(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func TestGuard_isFailure(t *testing.T) {
	guard := &Guard{cfg: Config{FailureStatuses: []int{401, 403}}}
	assert.True(t, guard.isFailure(401))
	assert.True(t, guard.isFailure(403))
	assert.False(t, guard.isFailure(200))
}

func TestGuard_block_default(t *testing.T) {
	guard := &Guard{cfg: Config{}}
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	guard.block(c, 10*time.Second)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["message"], "too many attempts")
	assert.Equal(t, float64(10), resp["retry_after"])
}

func TestGuard_block_negative_retry(t *testing.T) {
	guard := &Guard{cfg: Config{}}
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	guard.block(c, -5*time.Second) // Negative retry time
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["message"], "too many attempts")
	assert.Equal(t, float64(0), resp["retry_after"]) // Should be clamped to 0
}

func TestGuard_block_custom(t *testing.T) {
	called := false
	guard := &Guard{cfg: Config{OnBlocked: func(c *echo.Context, retryAfter time.Duration) {
		called = true
		c.String(444, "blocked")
	}}}
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	guard.block(c, 5*time.Second)
	assert.True(t, called)
	assert.Equal(t, 444, rec.Code)
	assert.Equal(t, "blocked", rec.Body.String())
}

func TestKeyByIPAndBodyField_JSON(t *testing.T) {
	body := map[string]string{"email": "User@EXAMPLE.com"}
	b, _ := json.Marshal(body)
	r := httptest.NewRequest("POST", "/", bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	r.RemoteAddr = "1.2.3.4:5678"
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(r, rec)
	key, err := KeyByIPAndBodyField("email")(c)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3.4:user@example.com", key)
}

func TestKeyByIPAndBodyField_JSON_no_field(t *testing.T) {
	body := map[string]string{"other": "value"}
	b, _ := json.Marshal(body)
	r := httptest.NewRequest("POST", "/", bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	r.RemoteAddr = "5.6.7.8:1234"
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(r, rec)
	key, err := KeyByIPAndBodyField("email")(c)
	assert.NoError(t, err)
	assert.Equal(t, "5.6.7.8", key)
}

func TestKeyByIPAndBodyField_Form(t *testing.T) {
	r := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("email=TestUser")))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.RemoteAddr = "9.8.7.6:4321"
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(r, rec)
	c.Request().ParseForm()
	c.Request().Form.Set("email", "TestUser")  // Use Form for Echo
	key, err := KeyByIPAndBodyField("email")(c)
	assert.NoError(t, err)
	assert.Equal(t, "9.8.7.6:testuser", key)
}

func TestKeyByIPAndBodyField_Error(t *testing.T) {
	r := httptest.NewRequest("POST", "/", &errReader{})
	r.Header.Set("Content-Type", "application/json")
	r.RemoteAddr = "1.1.1.1:1111"
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(r, rec)
	key, err := KeyByIPAndBodyField("email")(c)
	assert.NoError(t, err)
	assert.Equal(t, "1.1.1.1", key)
}

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("fail")
}

func (e *errReader) Close() error { return nil }

func TestNew_defaults(t *testing.T) {
	guard := New(Config{}, &MockService{}, func(*echo.Context) (string, error) { return "k", nil }, zap.NewNop().Sugar())
	assert.Equal(t, 5, guard.cfg.MaxAttempts)
	assert.Equal(t, time.Minute, guard.cfg.Window)
	assert.Equal(t, 1*time.Minute, guard.cfg.Lockout)
	assert.Equal(t, []int{401, 403}, guard.cfg.FailureStatuses)
}

func TestGuard_Middleware(t *testing.T) {
	mockSvc := &MockService{}
	logger := zap.NewNop().Sugar()
	key := "ip:user"
	cfg := Config{MaxAttempts: 2, Window: time.Minute, Lockout: 2 * time.Minute, FailureStatuses: []int{401}}
	guard := New(cfg, mockSvc, func(c *echo.Context) (string, error) { return key, nil }, logger)

	// Not locked, success (should call Reset)
	mockSvc.On("IsLocked", mock.Anything, key).Return(false, time.Time{}, nil)
	mockSvc.On("Reset", mock.Anything, key).Return(nil)
	e := echo.New()
	req := httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().WriteHeader(200)
	next := func(c *echo.Context) error { return nil }
	guard.Middleware()(next)(c)
	mockSvc.AssertCalled(t, "IsLocked", mock.Anything, key)
	mockSvc.AssertCalled(t, "Reset", mock.Anything, key)

	// Not locked, non-success status (should NOT call Reset)
	mockSvc = &MockService{}
	guard = New(cfg, mockSvc, func(c *echo.Context) (string, error) { return key, nil }, logger)
	mockSvc.On("IsLocked", mock.Anything, key).Return(false, time.Time{}, nil)
	// Note: NOT setting up Reset expectation
	e = echo.New()
	req = httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.Background())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Response().WriteHeader(400) // Non-success status
	guard.Middleware()(next)(c)
	mockSvc.AssertCalled(t, "IsLocked", mock.Anything, key)
	mockSvc.AssertNotCalled(t, "Reset") // Should NOT call Reset

	// Not locked, 500 status (should NOT call Reset)
	mockSvc = &MockService{}
	guard = New(cfg, mockSvc, func(c *echo.Context) (string, error) { return key, nil }, logger)
	mockSvc.On("IsLocked", mock.Anything, key).Return(false, time.Time{}, nil)
	// Note: NOT setting up Reset expectation
	e = echo.New()
	req = httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.Background())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Response().WriteHeader(500) // Non-success status
	guard.Middleware()(next)(c)
	mockSvc.AssertCalled(t, "IsLocked", mock.Anything, key)
	mockSvc.AssertNotCalled(t, "Reset") // Should NOT call Reset

	// Locked, should block
	mockSvc = &MockService{}
	guard = New(cfg, mockSvc, func(c *echo.Context) (string, error) { return key, nil }, logger)
	mockSvc.On("IsLocked", mock.Anything, key).Return(true, time.Now().Add(time.Minute), nil)
	e = echo.New()
	req = httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.Background())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	guard.Middleware()(next)(c)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// Locked but expired (race condition fix), should allow request
	mockSvc = &MockService{}
	guard = New(cfg, mockSvc, func(c *echo.Context) (string, error) { return key, nil }, logger)
	mockSvc.On("IsLocked", mock.Anything, key).Return(true, time.Now().Add(-time.Second), nil) // Expired lock
	e = echo.New()
	req = httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.Background())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Response().WriteHeader(200)
	guard.Middleware()(next)(c)
	assert.Equal(t, 200, rec.Code) // Should allow request to proceed

	// IsLocked error, should log and continue
	mockSvc = &MockService{}
	guard = New(cfg, mockSvc, func(c *echo.Context) (string, error) { return key, nil }, logger)
	mockSvc.On("IsLocked", mock.Anything, key).Return(false, time.Time{}, errors.New("fail"))
	e = echo.New()
	req = httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.Background())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	guard.Middleware()(next)(c)
	assert.Equal(t, 200, rec.Code)

	// Failure status, should call OnFailure
	mockSvc = &MockService{}
	guard = New(cfg, mockSvc, func(c *echo.Context) (string, error) { return key, nil }, logger)
	mockSvc.On("IsLocked", mock.Anything, key).Return(false, time.Time{}, nil)
	mockSvc.On("OnFailure", mock.Anything, key, mock.Anything, cfg.Window, cfg.MaxAttempts, cfg.Lockout).Return(true, time.Now().Add(cfg.Lockout), nil)
	e = echo.New()
	req = httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.Background())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Response().WriteHeader(401)
	guard.Middleware()(next)(c)
	mockSvc.AssertCalled(t, "OnFailure", mock.Anything, key, mock.Anything, cfg.Window, cfg.MaxAttempts, cfg.Lockout)
}

func TestNewGuard(t *testing.T) {
	// Test the NewGuard function from dig.go
	mockService := &MockService{}
	logger := zap.NewNop().Sugar()
	cfg := &config.Config{
		BruteforceMaxAttempts: 10,
		BruteforceWindow:      2 * time.Minute,
		BruteforceLockout:     30 * time.Minute,
	}

	guard := NewGuard(mockService, logger, cfg)
	assert.NotNil(t, guard)
	assert.Equal(t, 10, guard.cfg.MaxAttempts)
	assert.Equal(t, 2*time.Minute, guard.cfg.Window)
	assert.Equal(t, 30*time.Minute, guard.cfg.Lockout)
	assert.Equal(t, []int{401, 403}, guard.cfg.FailureStatuses)
}
