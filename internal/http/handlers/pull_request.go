package handlers

import (
	"github.com/Unitazavr/AvitoPR/internal/domain"
	"github.com/Unitazavr/AvitoPR/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PrHandler struct {
	prService service.PrService
}

func NewPrHandler(prService service.PrService) *PrHandler {
	return &PrHandler{
		prService: prService,
	}
}

// CreatePR - POST /pullRequests/create
func (h *PrHandler) CreatePR(c *gin.Context) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr := &domain.PullRequestShort{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
		Status:          domain.PRStatusOpen,
	}

	createdPR, err := h.prService.CreatePR(c.Request.Context(), pr)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": createdPR})
}

// MergePR - POST /pullRequest/merge
func (h *PrHandler) MergePR(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mergedPR, err := h.prService.MergePR(c.Request.Context(), req.PullRequestID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": mergedPR})
}

// ReassignPR - POST /pullRequests/reassign
func (h *PrHandler) ReassignPR(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedPR, replacedBy, err := h.prService.ReassignPR(c.Request.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          updatedPR,
		"replaced_by": replacedBy,
	})
}
