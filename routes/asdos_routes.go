package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func AsdosRoutes(r *gin.RouterGroup) {
	asdos := r.Group("/asdos")
	asdos.Use(IsKoordinatorMiddleware())
	{
		asdos.POST("/", controllers.CreateAsdos)
		asdos.GET("/", controllers.GetAllAsdos)
		asdos.GET("/:id", controllers.GetAsdosByID)
		asdos.PATCH("/:id", controllers.UpdateAsdos)
		asdos.DELETE("/:id", controllers.DeleteAsdos)
	}
}

