package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func RoomRoutes(r *gin.RouterGroup) {
	room := r.Group("/rooms")
	room.GET("/", controllers.GetAllRooms)
	room.Use(IsKoordinatorMiddleware())
	{
		room.POST("/", controllers.CreateRoom)
		room.GET("/:id", controllers.GetRoomByID)
		room.PATCH("/:id", controllers.UpdateRoom)
		room.DELETE("/:id", controllers.DeleteRoom)
	}
}
