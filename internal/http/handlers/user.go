package handlers

import (
	"github.com/Unitazavr/AvitoPR/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// SetIsActive - POST /users/setIsActive
func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// GetUserReviews - GET /users/getReview
func (h *UserHandler) GetUserReviews(c *gin.Context) {
	userID := c.Query("user_id")

	report, err := h.userService.GetUserReviews(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, report)
}
