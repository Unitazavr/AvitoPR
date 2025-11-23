package http

import (
	"errors"
	"github.com/Unitazavr/AvitoPR/internal/domain"
	"github.com/Unitazavr/AvitoPR/internal/http/handlers"
	"github.com/Unitazavr/AvitoPR/internal/repository"
	"github.com/Unitazavr/AvitoPR/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, userRepo repository.UserRepository, teamRepo repository.TeamRepository, prRepo repository.PrRepository) {
	userService := service.NewUserService(userRepo)
	teamService := service.NewTeamService(teamRepo)
	prService := service.NewPrService(prRepo)

	userHandler := handlers.NewUserHandler(userService)
	teamHandler := handlers.NewTeamHandler(teamService)
	prHandler := handlers.NewPrHandler(prService)

	router.Use(ErrorMiddleware())

	teamGroup := router.Group("/team")
	{
		teamGroup.POST("/add", teamHandler.CreateTeam)
		teamGroup.GET("/get", teamHandler.GetTeam)
	}

	usersGroup := router.Group("/users")
	{
		usersGroup.PATCH("/setIsActive", userHandler.SetIsActive)
		usersGroup.GET("/getReview", userHandler.GetUserReviews)
	}

	router.POST("/pullRequests/create", prHandler.CreatePR)
	router.PATCH("/pullRequest/merge", prHandler.MergePR)
	router.PATCH("/pullRequests/reassign", prHandler.ReassignPR)
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			statusCode := HandleError(err)

			var errResp *domain.ErrorResponse
			if errors.As(err, &errResp) {
				c.JSON(statusCode, errResp)
			} else {

				c.JSON(statusCode, domain.ErrorResponse{
					ErrorContent: domain.ErrorBody{
						Code:    domain.ErrUnknown,
						Message: err.Error(),
					},
				})
			}

			c.Abort()
		}
	}
}
