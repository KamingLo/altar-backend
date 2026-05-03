package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func KoorRoutes(r *gin.Engine) {
	koor := r.Group("/koor")
	koor.Use(AuthMiddleware(), IsKoordinatorMiddleware())
	{
		koor.POST("/", controllers.CreateKoordinator)
		koor.GET("/", controllers.GetAllKoordinator)
		koor.GET("/:id", controllers.GetKoordinatorByID)
		koor.PATCH("/:id", controllers.UpdateKoordinator)
		koor.DELETE("/:id", controllers.DeleteKoordinator)
	}
}
