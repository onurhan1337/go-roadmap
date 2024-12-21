package main

import (
	"bookstore-api/config"
	"bookstore-api/models"
	"log"
)

func main() {
	config.ConnectDatabase()

	err := config.DB.AutoMigrate(&models.Book{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	for _, book := range models.SeedBooks {
		result := config.DB.Create(&book)
		if result.Error != nil {
			log.Printf("Error seeding book %s: %v", book.Title, result.Error)
		} else {
			log.Printf("Seeded book: %s", book.Title)
		}
	}

	log.Println("Database seeding completed")
}