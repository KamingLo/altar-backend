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
		asdos.PATCH("/:id/deactivate", controllers.DeactivateAsdos)
		asdos.PATCH("/:id/activate", controllers.ActivateAsdos)
	}
}
