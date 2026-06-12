package engine

import (
	"context"
	"peekaping/internal/modules/certificate"
	"peekaping/internal/modules/healthcheck"
	"peekaping/internal/modules/proxy"
	"peekaping/internal/modules/shared"
	"time"
)

// CheckJob carries everything a worker goroutine needs to execute a health check.
type CheckJob struct {
	Monitor            *healthcheck.Monitor
	Proxy              *proxy.Model
	IsUnderMaintenance bool
	ScheduledAt        time.Time
	CheckCertExpiry    bool
	LastHeartbeat      *shared.HeartBeatModel
}

// CheckResult carries the outcome of a health check back to the ingester.
type CheckResult struct {
	MonitorID          string
	MonitorName        string
	MonitorType        string
	MonitorInterval    int
	MonitorTimeout     int
	MonitorMaxRetries  int
	MonitorRetryInt    int
	MonitorResendInt   int
	MonitorConfig      string
	Status             shared.MonitorStatus
	Message            string
	PingMs             int
	StartTime          time.Time
	EndTime            time.Time
	IsUnderMaintenance bool
	TLSInfo            *certificate.TLSInfo
	CheckCertExpiry    bool
}

// Queue is the job channel between scheduler and workers.
type Queue interface {
	Push(ctx context.Context, job CheckJob) error
	Pop(ctx context.Context) (CheckJob, error)
	Close()
}

// ResultQueue is the result channel between workers and the ingester.
type ResultQueue interface {
	Push(ctx context.Context, result CheckResult) error
	Pop(ctx context.Context) (CheckResult, error)
	Close()
}
