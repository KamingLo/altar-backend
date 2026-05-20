package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func ClassRoutes(r *gin.RouterGroup) {
	class := r.Group("/classes")
	class.Use(IsKoordinatorMiddleware())
	{
		class.POST("/", controllers.CreateClass)
		class.GET("/", controllers.GetAllClasses)
		class.GET("/:id", controllers.GetClassByID)
		class.PATCH("/:id", controllers.UpdateClass)
		class.DELETE("/:id", controllers.DeleteClass)
	}
}

