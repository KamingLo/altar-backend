package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func LecturerRoutes(r *gin.Engine) {
	lecturer := r.Group("/lecturers")
	{
		lecturer.POST("/", controllers.CreateLecturer)
		lecturer.GET("/", controllers.GetAllLecturers)
		lecturer.GET("/:id", controllers.GetLecturerByID)
		lecturer.PATCH("/:id", controllers.UpdateLecturer)
		lecturer.DELETE("/:id", controllers.DeleteLecturer)
	}
}

