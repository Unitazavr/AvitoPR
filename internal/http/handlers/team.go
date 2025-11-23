package handlers

import (
	"github.com/Unitazavr/AvitoPR/internal/domain"
	"github.com/Unitazavr/AvitoPR/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

// TeamHandler - обработчик для команд
type TeamHandler struct {
	teamService service.TeamService
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// CreateTeam - POST /team/add
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req struct {
		TeamName string              `json:"team_name"`
		Members  []domain.TeamMember `json:"members"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team := &domain.Team{
		TeamName: req.TeamName,
		Members:  req.Members,
	}

	createdTeam, err := h.teamService.CreateTeam(c.Request.Context(), team)
	if err != nil {

		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": createdTeam})
}

// GetTeam - GET /team/get
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")

	team, err := h.teamService.GetTeamByName(c.Request.Context(), teamName)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, team)
}
