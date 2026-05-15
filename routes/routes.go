package routes

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// r.Use(CORSMiddleware())
	// r.Use(RateLimitMiddleware())

	AuthRoutes(r)
	UserRoutes(r)
	AsdosRoutes(r)
	KoorRoutes(r)
	RoomRoutes(r)
	CourseRoutes(r)
	SemesterRoutes(r)
	ClassRoutes(r)
	LecturerRoutes(r)
	SessionRoutes(r)
	SubstituteSessionRoutes(r)

	return r
}


