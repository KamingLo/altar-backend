package routes

import (
	"altar/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		// Public Routes
		auth.GET("/google", controllers.GoogleLogin)
		auth.GET("/google/callback", controllers.GoogleCallback)
		auth.POST("/login", controllers.Login)
		auth.POST("/forgot-password", controllers.ForgotPassword)
		auth.POST("/reset-password", controllers.ResetPassword)

		// Koordinator Protected Routes
		koor := auth.Group("/")
		koor.Use(AuthMiddleware(), IsKoordinatorMiddleware())
		{
			koor.GET("/is-koor", controllers.CheckKoor)
		}

		// Asisten Dosen Protected Routes
		asdos := auth.Group("/")
		asdos.Use(AuthMiddleware(), IsAsdosMiddleware())
		{
			asdos.GET("/is-asdos", controllers.CheckAsdos)
		}

		// Private Routes (login needed)
		private := auth.Group("/")
		private.Use(AuthMiddleware())
		{
			private.GET("/me", controllers.GetMe)
			private.GET("/logout", controllers.Logout)
		}
	}
}
