package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func SemesterRoutes(r *gin.RouterGroup) {
	semester := r.Group("/semesters")
	{
		semester.POST("/", controllers.CreateSemester)
		semester.GET("/", controllers.GetAllSemesters)
		semester.GET("/:id", controllers.GetSemesterByID)
		semester.PATCH("/:id", controllers.UpdateSemester)
		semester.DELETE("/:id", controllers.DeleteSemester)
	}
}
