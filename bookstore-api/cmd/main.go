package main

import (
	"bookstore-api/config"
	_ "bookstore-api/docs"
	"bookstore-api/internal/router"
	"bookstore-api/models"
	"log"
)

// @title           Bookstore API
// @version         1.0
// @description     A simple bookstore API with CRUD operations
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http
func main() {
	config.ConnectDatabase()

	if err := config.DB.AutoMigrate(&models.Book{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	r := router.SetupRouter()
	log.Fatal(r.Run(":8080"))
}
