package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func PresensiRoutes(r *gin.RouterGroup) {
	presensi := r.Group("/presensi")
	{
		// Asdos Routes
		asdos := presensi.Group("/")
		asdos.Use(IsAsdosMiddleware())
		{
			asdos.GET("/me", controllers.GetAllMyPresensi)
			asdos.POST("/check-in", controllers.CheckIn)
			asdos.POST("/check-out", controllers.CheckOut)
			asdos.POST("/online", controllers.OnlineAttendance)
		}

		// Koordinator Routes
		koor := presensi.Group("/")
		koor.Use(IsKoordinatorMiddleware())
		{
			koor.GET("/", controllers.GetAllPresensi)
			koor.PATCH("/:id/verify", controllers.VerifyPresensi)
		}
	}
}
