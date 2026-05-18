package routes

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(CORSMiddleware())
	r.Use(RateLimitMiddleware())

	// Public and self-managed Auth routes
	AuthRoutes(r)

	// Protected routes group
	api := r.Group("/")
	api.Use(AuthMiddleware(), KioskBlockerMiddleware())
	{
		UserRoutes(api)
		AsdosRoutes(api)
		KoorRoutes(api)
		RoomRoutes(api)
		CourseRoutes(api)
		SemesterRoutes(api)
		ClassRoutes(api)
		LecturerRoutes(api)
		SessionRoutes(api)
		SubstituteSessionRoutes(api)
		PresensiRoutes(api)
	}

	return r
}


