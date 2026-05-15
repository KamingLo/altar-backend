package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func CourseRoutes(r *gin.Engine) {
	course := r.Group("/courses")
	course.Use(AuthMiddleware(), IsKoordinatorMiddleware())
	{
		course.POST("/", controllers.CreateCourse)
		course.GET("/", controllers.GetAllCourses)
		course.GET("/:id", controllers.GetCourseByID)
		course.PATCH("/:id", controllers.UpdateCourse)
		course.DELETE("/:id", controllers.DeleteCourse)
	}
}
