package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func ClassRoutes(r *gin.Engine) {
	class := r.Group("/classes")
	class.Use(AuthMiddleware(), IsKoordinatorMiddleware())
	{
		class.POST("/", controllers.CreateClass)
		class.GET("/", controllers.GetAllClasses)
		class.GET("/:id", controllers.GetClassByID)
		class.PATCH("/:id", controllers.UpdateClass)
		class.DELETE("/:id", controllers.DeleteClass)
	}
}

