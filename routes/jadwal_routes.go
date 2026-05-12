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
	}
}
