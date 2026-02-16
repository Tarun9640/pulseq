package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Tarun9640/pulseq/internal/service"
	"github.com/gin-gonic/gin"
)

// What Handler Does? (VERY SIMPLE)

// Handler = traffic police

// Accept request
// Validate
// Call service
// Return response

// Business logic NOT here.

// TaskHandler → HTTP requests ni handle chestundi
// Service ni call chesi response ivvadam its job
type TaskHandler struct {
	service *service.TaskService
}

// Dependency Injection
// Handler ki service ni inject chestham
func NewTaskHandler(s *service.TaskService) *TaskHandler {
	return &TaskHandler{
		service: s,
	}
}

// Client nundi vastunna request structure
// binding:"required" → field compulsory
type CreateTaskRequest struct {
	Type    string `json:"type" binding:"required"`
	Payload map[string]interface{}   `json:"payload" binding:"required"`
}

// POST /tasks endpoint
func (h *TaskHandler) CreateTask(c *gin.Context) {

	var req CreateTaskRequest //memory allocation 

	// ✅ Step 1:
	// Client JSON ni struct lo convert chestundi
	// Invalid JSON → 400 error
	if err := c.ShouldBindJSON(&req); err != nil {   //if you want to Write Data to func use &(it gives address to where) bindly
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// ✅ Step 2:
	// Raw payload ni bytes ga convert chestham
	// Because DB lo JSONB store cheyyali
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse payload",
		})
		return
	}

	// ✅ Step 3:
	// Business logic call
	// Handler logic rayadu — service ni call chestundi
	task, err := h.service.CreateTask(
		c.Request.Context(), //Prevents zombie queries if client disconnects.
		req.Type,
		payloadBytes,
	)

	// ✅ Step 4:
	// If DB/service fail → 500 error
	if err != nil {
		log.Println("create task error:", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create task",
		})
		return
	}

	var payload map[string]interface{}

	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		payload = nil
	}

	resp := TaskResponse{
		ID:        task.ID.String(),
		Type:      task.Type,
		Status:    task.Status,
		Payload:   payload,
		CreatedAt: task.CreatedAt.Time.String(),
	}


	// ✅ Step 5:
	// Success response
	c.JSON(http.StatusCreated, resp) //Never Return DB Models Directly DB struct ≠ API response.
}

//this is DTO(how api response should be)
type TaskResponse struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Status    string      `json:"status"`
	Payload   interface{} `json:"payload"`
	CreatedAt string      `json:"created_at"`
}


