package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func KoorRoutes(r *gin.RouterGroup) {
	koor := r.Group("/koor")
	koor.Use(IsKoordinatorMiddleware())
	{
		koor.POST("/", controllers.CreateKoordinator)
		koor.GET("/", controllers.GetAllKoordinator)
		koor.GET("/:id", controllers.GetKoordinatorByID)
		koor.PATCH("/:id", controllers.UpdateKoordinator)
		koor.DELETE("/:id", controllers.DeleteKoordinator)

		// Kiosk Mode & QR
		koor.POST("/kiosk/pin", controllers.SetKioskPIN)
		koor.POST("/kiosk/verify", controllers.VerifyKioskPIN)
		koor.POST("/kiosk/deactivate", controllers.DeactivateKiosk)
		koor.GET("/kiosk/generate-qr", controllers.GenerateQR)
	}
}
