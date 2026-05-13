package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func SessionRoutes(r *gin.Engine) {
	session := r.Group("/sessions")
	session.Use(AuthMiddleware(), IsKoordinatorMiddleware())
	{
		session.POST("/", controllers.CreateSession)
		session.GET("/", controllers.GetAllSessions)
		session.GET("/:id", controllers.GetSessionByID)
		session.PATCH("/:id", controllers.UpdateSession)
		session.DELETE("/:id", controllers.DeleteSession)
	}
}
