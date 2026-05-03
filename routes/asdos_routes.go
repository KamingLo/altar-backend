package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func AsdosRoutes(r *gin.Engine) {
	asdos := r.Group("/asdos")
	asdos.Use(AuthMiddleware(), IsKoordinatorMiddleware())
	{
		asdos.POST("/", controllers.CreateAsdos)
		asdos.GET("/", controllers.GetAllAsdos)
		asdos.GET("/:id", controllers.GetAsdosByID)
		asdos.PATCH("/:id", controllers.UpdateAsdos)
		asdos.DELETE("/:id", controllers.DeleteAsdos)
	}
}
