# Peekaping server

## Docs
http://localhost:8034/swagger/index.html

## HTTP Framework
The API server uses Echo v5 (github.com/labstack/echo/v5) as the HTTP framework (migrated from Gin as part of change migrar-api-echo).
- All routes/controllers use `*echo.Context` and Echo idioms.
- Module generator templates (`templates/module/*.tmpl`) now emit native Echo code.
- See `internal/server.go` for setup and `cmd/api/main.go` for startup.
