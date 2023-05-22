package routes

import (
	"ecommerce/controllers"
	"ecommerce/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	// Public routes
	incomingRoutes.POST("/users/register", controllers.Register)
	incomingRoutes.POST("/users/login", controllers.Login)

	// Routes that require authentication
	auth := incomingRoutes.Group("/users")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/:id", controllers.GetUser)
	auth.PATCH("/update/:id", controllers.UpdateUser)
	auth.DELETE("/delete/:id", controllers.DeleteUser)
	auth.POST("/users/logout", controllers.LogoutUser)
	{

		// Routes that require admin role
		admin := auth.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/users", controllers.GetUsers)
			admin.GET("/:id", controllers.GetUser)
			admin.PATCH("/update/:id", controllers.UpdateUser)
			admin.DELETE("/:id", controllers.DeleteUser)
			admin.GET("/webtracking/count", controllers.GetWebTrackingCount)
			admin.GET("/related_product/:id", controllers.GetProductDetails)
		}
	}
}
