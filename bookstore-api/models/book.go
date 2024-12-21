package models

import "gorm.io/gorm"

// Book represents a book in the store
// @Description Book information with ISBN and price details
type Book struct {
	gorm.Model
	// ISBN is the unique identifier for the book
	ISBN      string  `json:"isbn" gorm:"uniqueIndex" binding:"required,isbn" example:"978-1503261969"`
	// Title of the book
	Title     string  `json:"title" binding:"required,min=1,max=100" example:"Emma"`
	// Author of the book
	Author    string  `json:"author" binding:"required,min=1,max=100" example:"Jane Austen"`
	// Publisher of the book
	Publisher string  `json:"publisher" binding:"required,min=1,max=100" example:"Wilder Publications"`
	// Price of the book in USD
	Price     float64 `json:"price" binding:"required,gt=0" example:"9.99"`
}

// BookResponse represents the response structure for a book
type BookResponse struct {
	ID        uint    `json:"id" example:"1"`
	CreatedAt string  `json:"created_at" example:"2024-12-21T14:30:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2024-12-21T14:30:00Z"`
	ISBN      string  `json:"isbn" example:"978-1503261969"`
	Title     string  `json:"title" example:"Emma"`
	Author    string  `json:"author" example:"Jane Austen"`
	Publisher string  `json:"publisher" example:"Wilder Publications"`
	Price     float64 `json:"price" example:"9.99"`
}

var SeedBooks = []Book{
	{ISBN: "978-1503261969", Title: "Emma", Author: "Jane Austen", Publisher: "Wilder Publications", Price: 9.95},
	{ISBN: "978-1505255607", Title: "The Time Machine", Author: "H. G. Wells", Publisher: "CreateSpace Independent Publishing Platform", Price: 5.99},
	{ISBN: "979-8601301471", Title: "Think and Grow Rich", Author: "Napoleon Hill", Publisher: "Independently published", Price: 7.95},
	{ISBN: "978-1503379640", Title: "The Prince", Author: "Niccolò Machiavelli", Publisher: "CreateSpace Independent Publishing Platform", Price: 6.99},
}