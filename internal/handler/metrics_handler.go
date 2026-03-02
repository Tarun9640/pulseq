package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Tarun9640/pulseq/internal/queue"
	"github.com/Tarun9640/pulseq/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type MetricsHandler struct {
	redisClient *redis.Client
	logger      *slog.Logger
	manager     *worker.WorkerManager
}



func NewMetricsHandler(redisClient *redis.Client, manager *worker.WorkerManager, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{
		redisClient: redisClient,
		manager: manager,
		logger: logger,
	}
}

func (h *MetricsHandler) GetMetrics(c *gin.Context) {

	ctx := context.Background()

	queueDepth, _ := h.redisClient.LLen(ctx, queue.TaskQueueName).Result()

	processing, _ := h.redisClient.LLen(ctx, queue.ProcessingQueueName).Result()

	delayQueue, _ := h.redisClient.ZCard(ctx, queue.DelayQueue).Result()

	dlqSize, _ := h.redisClient.LLen(ctx, queue.DeadLetterQueue).Result()

	workers := h.manager.GetWorkerCount()

	h.logger.Info("metrics requested")

	c.JSON(http.StatusOK, gin.H{
		"queue_depth": queueDepth,
		"processing_jobs": processing,
		"delay_queue": delayQueue,
		"dlq_size": dlqSize,
		"workers": workers,
	})
}