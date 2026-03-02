package handler

import (
	"context"
	"net/http"

	"github.com/Tarun9640/pulseq/internal/worker"
	"github.com/gin-gonic/gin"
)

type WorkerHandler struct {
	manager *worker.WorkerManager
}

func NewWorkerHandler(manager *worker.WorkerManager) *WorkerHandler{
	return &WorkerHandler{
		manager: manager,
	}
}

func (h *WorkerHandler) GetWorkers(c *gin.Context) {
	
	stats, err := h.manager.GetWorkerStats(context.Background())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "failed to get worker stats",
		})
		return
	}
	c.JSON(http.StatusOK, stats)
}