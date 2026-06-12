package engine

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"peekaping/internal/modules/heartbeat"
	"peekaping/internal/modules/maintenance"
	"peekaping/internal/modules/monitor"
	"peekaping/internal/modules/proxy"
	"peekaping/internal/modules/shared"

	"go.uber.org/zap"
)

// Scheduler polls the database for due monitors and pushes CheckJobs to the queue.
type Scheduler struct {
	monitorService     monitor.Service
	proxyService       proxy.Service
	maintenanceService maintenance.Service
	heartbeatService   heartbeat.Service
	queue              Queue
	cfg                Config
	logger             *zap.SugaredLogger

	mu         sync.Mutex
	nextRunAt  map[string]time.Time // monitorID -> next scheduled time
}

// NewScheduler creates a Scheduler.
func NewScheduler(
	monitorService monitor.Service,
	proxyService proxy.Service,
	maintenanceService maintenance.Service,
	heartbeatService heartbeat.Service,
	queue Queue,
	cfg Config,
	logger *zap.SugaredLogger,
) *Scheduler {
	return &Scheduler{
		monitorService:     monitorService,
		proxyService:       proxyService,
		maintenanceService: maintenanceService,
		heartbeatService:   heartbeatService,
		queue:              queue,
		cfg:                cfg,
		logger:             logger.With("component", "engine.scheduler"),
		nextRunAt:          make(map[string]time.Time),
	}
}

// Start runs the scheduler loop until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("Scheduler started")
	ticker := time.NewTicker(s.cfg.SchedulerInterval)
	defer ticker.Stop()

	// Run immediately on first start.
	s.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// tick finds all active monitors that are due and enqueues them.
func (s *Scheduler) tick(ctx context.Context) {
	now := time.Now().UTC()

	const pageSize = 100
	page := 0
	activeIDs := make(map[string]bool)

	for {
		monitors, err := s.monitorService.FindActivePaginated(ctx, page, pageSize)
		if err != nil {
			s.logger.Errorw("Failed to fetch active monitors", "error", err)
			return
		}
		if len(monitors) == 0 {
			break
		}

		for _, mon := range monitors {
			if mon.Interval <= 0 {
				continue
			}
			activeIDs[mon.ID] = true

			s.mu.Lock()
			next, scheduled := s.nextRunAt[mon.ID]
			s.mu.Unlock()

			if scheduled && now.Before(next) {
				continue // not due yet
			}

			// Build and enqueue the job.
			job, err := s.buildJob(ctx, mon, now)
			if err != nil {
				s.logger.Errorw("Failed to build check job", "monitor_id", mon.ID, "error", err)
				continue
			}

			if err := s.queue.Push(ctx, job); err != nil {
				s.logger.Errorw("Failed to push check job", "monitor_id", mon.ID, "error", err)
				continue
			}

			s.mu.Lock()
			s.nextRunAt[mon.ID] = now.Add(time.Duration(mon.Interval) * time.Second)
			s.mu.Unlock()

			s.logger.Debugw("Enqueued check job", "monitor_id", mon.ID, "monitor_name", mon.Name)
		}

		page++
		if len(monitors) < pageSize {
			break
		}
	}

	// Purge nextRunAt entries for monitors that are no longer active.
	s.mu.Lock()
	for id := range s.nextRunAt {
		if !activeIDs[id] {
			delete(s.nextRunAt, id)
		}
	}
	s.mu.Unlock()
}

// buildJob assembles a CheckJob for the given monitor.
func (s *Scheduler) buildJob(ctx context.Context, mon *shared.Monitor, now time.Time) (CheckJob, error) {
	isUnderMaintenance, err := s.checkMaintenance(ctx, mon.ID)
	if err != nil {
		s.logger.Warnw("Maintenance check failed, proceeding without maintenance status",
			"monitor_id", mon.ID, "error", err)
	}

	var proxyModel *proxy.Model
	if mon.ProxyId != "" {
		p, err := s.proxyService.FindByID(ctx, mon.ProxyId)
		if err != nil {
			s.logger.Warnw("Failed to fetch proxy, continuing without it",
				"monitor_id", mon.ID, "proxy_id", mon.ProxyId, "error", err)
		} else {
			proxyModel = p
		}
	}

	checkCertExpiry := false
	monType := strings.ToLower(mon.Type)
	if strings.HasPrefix(monType, "http") || monType == "tcp" {
		if mon.Config != "" {
			var cfg struct {
				CheckCertExpiry bool `json:"check_cert_expiry"`
			}
			if err := json.Unmarshal([]byte(mon.Config), &cfg); err == nil {
				checkCertExpiry = cfg.CheckCertExpiry
			}
		}
	}

	var lastHeartbeat *shared.HeartBeatModel
	if mon.Type == "push" {
		beats, err := s.heartbeatService.FindByMonitorIDPaginated(ctx, mon.ID, 1, 0, nil, false)
		if err != nil {
			s.logger.Warnw("Failed to fetch last heartbeat for push monitor", "monitor_id", mon.ID, "error", err)
		} else if len(beats) > 0 {
			lastHeartbeat = beats[0]
		}
	}

	return CheckJob{
		Monitor:            mon,
		Proxy:              proxyModel,
		IsUnderMaintenance: isUnderMaintenance,
		ScheduledAt:        now,
		CheckCertExpiry:    checkCertExpiry,
		LastHeartbeat:      lastHeartbeat,
	}, nil
}

func (s *Scheduler) checkMaintenance(ctx context.Context, monitorID string) (bool, error) {
	maintenances, err := s.maintenanceService.GetMaintenancesByMonitorID(ctx, monitorID)
	if err != nil {
		return false, err
	}
	for _, m := range maintenances {
		under, err := s.maintenanceService.IsUnderMaintenance(ctx, m)
		if err != nil {
			s.logger.Warnw("Failed to check maintenance status", "maintenance_id", m.ID, "error", err)
			continue
		}
		if under {
			return true, nil
		}
	}
	return false, nil
}
