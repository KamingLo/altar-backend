package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	user := r.Group("/users")
	user.Use(AuthMiddleware(), IsKoordinatorMiddleware())
	{
		// Asisten Dosen CRUD
		user.POST("/asdos", controllers.CreateAsdos)
		user.GET("/asdos", controllers.GetAllAsdos)
		user.GET("/asdos/:id", controllers.GetAsdosByID)
		user.PATCH("/asdos/:id", controllers.UpdateAsdos)
		user.DELETE("/asdos/:id", controllers.DeleteAsdos)

		// Koordinator CRUD
		user.POST("/koor", controllers.CreateKoordinator)
		user.GET("/koor", controllers.GetAllKoordinator)
		user.GET("/koor/:id", controllers.GetKoordinatorByID)
		user.PATCH("/koor/:id", controllers.UpdateKoordinator)
		user.DELETE("/koor/:id", controllers.DeleteKoordinator)
	}
}
