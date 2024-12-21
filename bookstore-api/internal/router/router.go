package router

import (
	"bookstore-api/controllers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	// Health check
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		books := v1.Group("/books")
		{
			books.GET("", controllers.GetBooks)
			books.GET("/:isbn", controllers.GetBook)
			books.POST("", controllers.CreateBook)
			books.PUT("/:isbn", controllers.UpdateBook)
			books.DELETE("/:isbn", controllers.DeleteBook)
		}
	}

	return router
}