package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	user := r.Group("/users")
	user.Use(AuthMiddleware(), IsKoordinatorMiddleware())
	{
		user.POST("/", controllers.CreateUser)
		user.GET("/", controllers.GetAllUsers)
		user.GET("/:id", controllers.GetUserByID)
		user.PATCH("/:id", controllers.UpdateUser)
		user.DELETE("/:id", controllers.DeleteUser)
	}
}
