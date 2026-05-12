package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func RoomRoutes(r *gin.Engine) {
	room := r.Group("/rooms")
	{
		room.POST("/", controllers.CreateRoom)
		room.GET("/", controllers.GetAllRooms)
		room.GET("/:id", controllers.GetRoomByID)
		room.PATCH("/:id", controllers.UpdateRoom)
		room.DELETE("/:id", controllers.DeleteRoom)
	}
}
