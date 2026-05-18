package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func SubstituteSessionRoutes(r *gin.RouterGroup) {
	sub := r.Group("/substitute-sessions")
	{
		// Asdos: submit a new substitute session request
		sub.POST("/", controllers.CreateSubstituteSession)
		sub.GET("/:id", controllers.GetSubstituteSessionByID)
		sub.DELETE("/:id", controllers.DeleteSubstituteSession)

		// Koordinator: review the queue and update status
		sub.GET("/", IsKoordinatorMiddleware(), controllers.GetAllSubstituteSessions)
		sub.PATCH("/:id/status", IsKoordinatorMiddleware(), controllers.UpdateSubstituteStatus)
	}

	// Jadwal (schedule) endpoints — all require authentication
	jadwal := r.Group("/jadwal")
	{
		// Global view: all sessions for a semester (any authenticated user)
		jadwal.GET("/sessions", controllers.GetScheduleTimeline)

		// Personal view: sessions for the requesting asdos only
		jadwal.GET("/my-sessions", IsAsdosMiddleware(), controllers.GetMyScheduleTimeline)
	}
}
