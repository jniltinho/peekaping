# API Presentation Layer (HTTP Delivery)

This specification captures the invariant requirements for the public HTTP API of Peekaping. The migration from Gin to Echo is an **internal implementation detail** and must not alter observable behavior.

## ADDED Requirements

### Requirement: Public HTTP API contract is stable across framework changes

The system SHALL preserve the exact external HTTP API surface (paths, methods, request/response shapes, status codes, authentication mechanisms, error message formats, and header behavior) when the underlying web framework implementation is changed.

#### Scenario: Successful monitor CRUD after framework migration
- **WHEN** a client performs GET/POST/PUT/PATCH/DELETE on `/api/v1/monitors*` (and all other resource groups) using valid JWT or X-API-Key
- **THEN** it receives identical status codes, response envelopes (`{ "message": "...", "data": ... }`), and error shapes as before the migration.

#### Scenario: Push monitor ingestion
- **WHEN** an external system POSTs or GETs the `/api/v1/push/:token` endpoint with the documented query parameters
- **THEN** the heartbeat is enqueued and the client receives the same success/failure responses.

#### Scenario: Authentication and brute-force protection
- **WHEN** repeated failed logins occur
- **THEN** the account is locked with the same retry-after semantics and status codes, independent of the router implementation.

#### Scenario: Real-time updates via socket.io compatibility layer
- **WHEN** the web UI connects via the `/socket.io/` endpoints
- **THEN** realtime heartbeat and status events continue to flow exactly as before.

## MODIFIED Requirements

(none — the migration changes only the internal framework; no public requirement is modified)

## REMOVED Requirements

(none)

## Notes for Implementers

- All Swagger annotations on controller methods must remain valid and produce an identical OpenAPI document (after regeneration).
- The standardized response helpers (`utils.NewSuccessResponse`, `NewFailResponse`, `ApiResponse[T]`) are the contract — they must continue to be used.
- Any new Echo-specific code in the HTTP layer (controllers, routes, middleware) is considered presentation-layer implementation and is out of scope for behavioral specification changes.