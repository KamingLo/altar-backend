package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)


func SessionRoutes(r *gin.Engine) {
	session := r.Group("/sessions")

	// Global / Open routes
	session.GET("/", controllers.GetAllSessions)

	// Asdos specific routes
	session.GET("/me", AuthMiddleware(), IsAsdosMiddleware(), controllers.GetMySession)

	// Detail route (put after static routes)
	// session.GET("/:id", controllers.GetSessionByID)

	// Koordinator routes
	koor := session.Group("/")
	koor.Use(AuthMiddleware(), IsKoordinatorMiddleware())
	{
		koor.POST("/", controllers.CreateSession)
		koor.PATCH("/:id", controllers.UpdateSession)
		koor.DELETE("/:id", controllers.DeleteSession)
	}
}



