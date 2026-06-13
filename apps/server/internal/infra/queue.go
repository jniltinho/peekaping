package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"peekaping/internal/config"
	"peekaping/internal/modules/queue"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// ProvideAsynqClient creates an asynq.Client connected to Redis via cfg.
func ProvideAsynqClient(
	cfg *config.Config,
	logger *zap.SugaredLogger,
) (*asynq.Client, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	client := asynq.NewClient(redisOpt)
	logger.Info("Successfully created Asynq client")
	return client, nil
}

// ProvideAsynqServer creates an asynq.Server with four strict-priority queues
// (critical=6 > healthcheck=5 > default=3 > low=1) and cfg.QueueConcurrency workers.
func ProvideAsynqServer(
	cfg *config.Config,
	logger *zap.SugaredLogger,
) (*asynq.Server, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	serverCfg := asynq.Config{
		Concurrency: cfg.QueueConcurrency,
		Queues: map[string]int{
			"critical":    6,
			"healthcheck": 5,
			"default":     3,
			"low":         1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Errorw("Task processing failed",
				"type", task.Type(),
				"payload", string(task.Payload()),
				"error", err,
			)
		}),
		StrictPriority: true,
		Logger:         NewAsynqLogger(logger),
	}

	server := asynq.NewServer(redisOpt, serverCfg)
	logger.Info("Successfully created Asynq server")
	return server, nil
}

// AsynqLogger adapts zap.SugaredLogger to the asynq.Logger interface.
type AsynqLogger struct {
	logger *zap.SugaredLogger
}

// NewAsynqLogger wraps logger in an AsynqLogger.
func NewAsynqLogger(logger *zap.SugaredLogger) *AsynqLogger {
	return &AsynqLogger{logger: logger}
}

func (l *AsynqLogger) Debug(args ...any) { l.logger.Debug(args...) }
func (l *AsynqLogger) Info(args ...any)  { l.logger.Info(args...) }
func (l *AsynqLogger) Warn(args ...any)  { l.logger.Warn(args...) }
func (l *AsynqLogger) Error(args ...any) { l.logger.Error(args...) }
func (l *AsynqLogger) Fatal(args ...any) { l.logger.Fatal(args...) }

// ProvideAsynqInspector creates an asynq.Inspector for queue introspection.
func ProvideAsynqInspector(
	cfg *config.Config,
	logger *zap.SugaredLogger,
) (*asynq.Inspector, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	inspector := asynq.NewInspector(redisOpt)
	logger.Info("Successfully created Asynq inspector")
	return inspector, nil
}

// ProvideAsynqScheduler creates an asynq.Scheduler using cfg.Timezone for
// cron expression evaluation.
func ProvideAsynqScheduler(
	cfg *config.Config,
	logger *zap.SugaredLogger,
) (*asynq.Scheduler, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		logger.Warnw("Failed to load timezone, using UTC", "timezone", cfg.Timezone, "error", err)
		location = time.UTC
	}

	schedulerCfg := &asynq.SchedulerOpts{
		Location: location,
		Logger:   NewAsynqLogger(logger),
		EnqueueErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
			logger.Errorw("Failed to enqueue scheduled task",
				"type", task.Type(),
				"error", err,
			)
		},
	}

	scheduler := asynq.NewScheduler(redisOpt, schedulerCfg)
	logger.Info("Successfully created Asynq scheduler")
	return scheduler, nil
}

// queueServiceImpl implements queue.Service using asynq.
type queueServiceImpl struct {
	client    *asynq.Client
	inspector *asynq.Inspector
	logger    *zap.SugaredLogger
}

// ProvideQueueService wraps an asynq.Client and asynq.Inspector as a queue.Service
// for use in the DI container.
func ProvideQueueService(
	client *asynq.Client,
	inspector *asynq.Inspector,
	logger *zap.SugaredLogger,
) queue.Service {
	return &queueServiceImpl{
		client:    client,
		inspector: inspector,
		logger:    logger.Named("[queue-service]"),
	}
}

func (s *queueServiceImpl) Enqueue(ctx context.Context, taskType string, payload any, opts *queue.EnqueueOptions) (*queue.TaskInfo, error) {
	if opts == nil {
		opts = queue.DefaultEnqueueOptions()
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Errorw("Failed to marshal payload", "task_type", taskType, "error", err)
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(taskType, payloadBytes)
	asynqOpts := buildAsynqOptions(opts)

	var info *asynq.TaskInfo
	if opts.ProcessAt != nil {
		info, err = s.client.Enqueue(task, append(asynqOpts, asynq.ProcessAt(*opts.ProcessAt))...)
	} else if opts.ProcessIn != nil {
		info, err = s.client.Enqueue(task, append(asynqOpts, asynq.ProcessIn(*opts.ProcessIn))...)
	} else {
		info, err = s.client.Enqueue(task, asynqOpts...)
	}

	if err != nil {
		s.logger.Errorw("Failed to enqueue task", "task_type", taskType, "error", err)
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return convertTaskInfo(info), nil
}

func (s *queueServiceImpl) EnqueueUnique(
	_ context.Context,
	taskType string,
	payload any,
	uniqueKey string,
	ttl time.Duration,
	opts *queue.EnqueueOptions,
) (*queue.TaskInfo, error) {
	if opts == nil {
		opts = queue.DefaultEnqueueOptions()
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Errorw("Failed to marshal payload", "task_type", taskType, "error", err)
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(taskType, payloadBytes)
	asynqOpts := buildAsynqOptions(opts)
	asynqOpts = append(asynqOpts, asynq.Unique(ttl))

	if opts.TaskID == "" {
		opts.TaskID = uniqueKey
	}
	asynqOpts = append(asynqOpts, asynq.TaskID(opts.TaskID))

	var info *asynq.TaskInfo
	if opts.ProcessAt != nil {
		info, err = s.client.Enqueue(task, append(asynqOpts, asynq.ProcessAt(*opts.ProcessAt))...)
	} else if opts.ProcessIn != nil {
		info, err = s.client.Enqueue(task, append(asynqOpts, asynq.ProcessIn(*opts.ProcessIn))...)
	} else {
		info, err = s.client.Enqueue(task, asynqOpts...)
	}

	if err != nil {
		errMsg := err.Error()
		// asynq surfaces deduplication collisions as non-fatal errors; log at Debug.
		if strings.Contains(errMsg, "task ID conflicts") ||
			strings.Contains(errMsg, "duplicated") ||
			strings.Contains(errMsg, "already exists") {
			s.logger.Debugw("Task already queued (duplicate prevented by unique constraint)",
				"task_type", taskType,
				"unique_key", uniqueKey)
			return nil, fmt.Errorf("task already exists: %w", err)
		}
		s.logger.Errorw("Failed to enqueue unique task", "task_type", taskType, "unique_key", uniqueKey, "error", err)
		return nil, fmt.Errorf("failed to enqueue unique task: %w", err)
	}

	return convertTaskInfo(info), nil
}

func (s *queueServiceImpl) GetQueueInfo(_ context.Context, queueName string) (*queue.QueueInfo, error) {
	info, err := s.inspector.GetQueueInfo(queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue info: %w", err)
	}
	return convertQueueInfo(info), nil
}

func (s *queueServiceImpl) ListQueues(_ context.Context) ([]*queue.QueueInfo, error) {
	queues, err := s.inspector.Queues()
	if err != nil {
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	var queueInfos []*queue.QueueInfo
	for _, queueName := range queues {
		info, err := s.inspector.GetQueueInfo(queueName)
		if err != nil {
			s.logger.Warnw("Failed to get queue info", "queue", queueName, "error", err)
			continue
		}
		queueInfos = append(queueInfos, convertQueueInfo(info))
	}

	return queueInfos, nil
}

func (s *queueServiceImpl) GetTaskInfo(_ context.Context, queueName, taskID string) (*queue.TaskInfo, error) {
	info, err := s.inspector.GetTaskInfo(queueName, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task info: %w", err)
	}
	return convertTaskInfo(info), nil
}

func (s *queueServiceImpl) DeleteTask(_ context.Context, queueName, taskID string) error {
	if err := s.inspector.DeleteTask(queueName, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	s.logger.Infow("Task deleted", "queue", queueName, "task_id", taskID)
	return nil
}

func (s *queueServiceImpl) CancelTask(_ context.Context, taskID string) error {
	if err := s.inspector.CancelProcessing(taskID); err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}
	s.logger.Infow("Task cancelled", "task_id", taskID)
	return nil
}

func (s *queueServiceImpl) PauseQueue(_ context.Context, queueName string) error {
	if err := s.inspector.PauseQueue(queueName); err != nil {
		return fmt.Errorf("failed to pause queue: %w", err)
	}
	s.logger.Infow("Queue paused", "queue", queueName)
	return nil
}

func (s *queueServiceImpl) UnpauseQueue(_ context.Context, queueName string) error {
	if err := s.inspector.UnpauseQueue(queueName); err != nil {
		return fmt.Errorf("failed to unpause queue: %w", err)
	}
	s.logger.Infow("Queue unpaused", "queue", queueName)
	return nil
}

func (s *queueServiceImpl) ListPendingTasks(_ context.Context, queueName string, pageSize, pageNum int) ([]*queue.TaskInfo, error) {
	tasks, err := s.inspector.ListPendingTasks(queueName, asynq.PageSize(pageSize), asynq.Page(pageNum))
	if err != nil {
		return nil, fmt.Errorf("failed to list pending tasks: %w", err)
	}
	return convertTaskInfoList(tasks), nil
}

func (s *queueServiceImpl) ListActiveTasks(_ context.Context, queueName string, pageSize, pageNum int) ([]*queue.TaskInfo, error) {
	tasks, err := s.inspector.ListActiveTasks(queueName, asynq.PageSize(pageSize), asynq.Page(pageNum))
	if err != nil {
		return nil, fmt.Errorf("failed to list active tasks: %w", err)
	}
	return convertTaskInfoList(tasks), nil
}

func (s *queueServiceImpl) ListScheduledTasks(_ context.Context, queueName string, pageSize, pageNum int) ([]*queue.TaskInfo, error) {
	tasks, err := s.inspector.ListScheduledTasks(queueName, asynq.PageSize(pageSize), asynq.Page(pageNum))
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks: %w", err)
	}
	return convertTaskInfoList(tasks), nil
}

func (s *queueServiceImpl) Close() error {
	if err := s.client.Close(); err != nil {
		return err
	}
	if err := s.inspector.Close(); err != nil {
		return err
	}
	s.logger.Info("Queue service closed")
	return nil
}

// buildAsynqOptions converts generic EnqueueOptions to the asynq option slice.
func buildAsynqOptions(opts *queue.EnqueueOptions) []asynq.Option {
	asynqOpts := []asynq.Option{
		asynq.Queue(opts.Queue),
		asynq.MaxRetry(opts.MaxRetry),
		asynq.Timeout(opts.Timeout),
		asynq.Retention(opts.Retention),
	}
	if opts.TaskID != "" {
		asynqOpts = append(asynqOpts, asynq.TaskID(opts.TaskID))
	}
	if opts.Deadline != nil {
		asynqOpts = append(asynqOpts, asynq.Deadline(*opts.Deadline))
	}
	return asynqOpts
}

func convertTaskInfo(info *asynq.TaskInfo) *queue.TaskInfo {
	if info == nil {
		return nil
	}
	return &queue.TaskInfo{
		ID:            info.ID,
		Queue:         info.Queue,
		Type:          info.Type,
		Payload:       info.Payload,
		State:         info.State.String(),
		MaxRetry:      info.MaxRetry,
		Retried:       info.Retried,
		LastErr:       info.LastErr,
		LastFailedAt:  info.LastFailedAt,
		NextProcessAt: info.NextProcessAt,
	}
}

func convertTaskInfoList(infos []*asynq.TaskInfo) []*queue.TaskInfo {
	result := make([]*queue.TaskInfo, 0, len(infos))
	for _, info := range infos {
		result = append(result, convertTaskInfo(info))
	}
	return result
}

func convertQueueInfo(info *asynq.QueueInfo) *queue.QueueInfo {
	if info == nil {
		return nil
	}
	return &queue.QueueInfo{
		Queue:     info.Queue,
		Size:      info.Size,
		Pending:   info.Pending,
		Active:    info.Active,
		Scheduled: info.Scheduled,
		Retry:     info.Retry,
		Archived:  info.Archived,
		Completed: info.Completed,
		Processed: info.Processed,
		Failed:    info.Failed,
		Paused:    info.Paused,
		Timestamp: info.Timestamp,
	}
}
