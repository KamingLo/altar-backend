package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func SubstituteSessionRoutes(r *gin.Engine) {
	sub := r.Group("/substitute-sessions")
	sub.Use(AuthMiddleware())
	{
		// Asdos: submit a new substitute session request
		sub.POST("/", IsAsdosMiddleware(), controllers.CreateSubstituteSession)

		// Koordinator: review the queue and update status
		sub.GET("/", IsKoordinatorMiddleware(), controllers.GetAllSubstituteSessions)
		sub.PATCH("/:id/status", IsKoordinatorMiddleware(), controllers.UpdateSubstituteStatus)
	}
}
