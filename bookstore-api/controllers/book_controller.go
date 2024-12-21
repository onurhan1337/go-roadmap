package controllers

import (
	"bookstore-api/config"
	"bookstore-api/models"
	"bookstore-api/pkg/validator"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var bookValidator = validator.NewBookValidator()

// @Summary Get all books
// @Description Get a list of all books
// @Tags books
// @Accept json
// @Produce json
// @Success 200 {array} models.BookResponse
// @Failure 500 {object} map[string]string
// @Router /books [get]
func GetBooks(c *gin.Context) {
	var books []models.Book
	result := config.DB.Find(&books)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching books"})
		return
	}

	var response []models.BookResponse
	for _, book := range books {
		response = append(response, models.BookResponse{
			ID:        book.ID,
			CreatedAt: book.CreatedAt.Format(time.RFC3339),
			UpdatedAt: book.UpdatedAt.Format(time.RFC3339),
			ISBN:      book.ISBN,
			Title:     book.Title,
			Author:    book.Author,
			Publisher: book.Publisher,
			Price:     book.Price,
		})
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get a book by ISBN
// @Description Get a book's details by its ISBN
// @Tags books
// @Accept json
// @Produce json
// @Param isbn path string true "Book ISBN"
// @Success 200 {object} models.BookResponse
// @Failure 404 {object} map[string]string
// @Router /books/{isbn} [get]
func GetBook(c *gin.Context) {
	isbn := c.Param("isbn")
	var book models.Book

	result := config.DB.Where("isbn = ?", isbn).First(&book)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}

	response := models.BookResponse{
		ID:        book.ID,
		CreatedAt: book.CreatedAt.Format(time.RFC3339),
		UpdatedAt: book.UpdatedAt.Format(time.RFC3339),
		ISBN:      book.ISBN,
		Title:     book.Title,
		Author:    book.Author,
		Publisher: book.Publisher,
		Price:     book.Price,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Create a new book
// @Description Add a new book to the store
// @Tags books
// @Accept json
// @Produce json
// @Param book body models.Book true "Book object"
// @Success 201 {object} models.BookResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /books [post]
func CreateBook(c *gin.Context) {
	var newBook models.Book

	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := bookValidator.Validate(newBook); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	result := config.DB.Create(&newBook)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating book"})
		return
	}

	response := models.BookResponse{
		ID:        newBook.ID,
		CreatedAt: newBook.CreatedAt.Format(time.RFC3339),
		UpdatedAt: newBook.UpdatedAt.Format(time.RFC3339),
		ISBN:      newBook.ISBN,
		Title:     newBook.Title,
		Author:    newBook.Author,
		Publisher: newBook.Publisher,
		Price:     newBook.Price,
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary Update a book
// @Description Update a book's details by ISBN
// @Tags books
// @Accept json
// @Produce json
// @Param isbn path string true "Book ISBN"
// @Param book body models.Book true "Book object"
// @Success 200 {object} models.BookResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books/{isbn} [put]
func UpdateBook(c *gin.Context) {
	isbn := c.Param("isbn")
	var book models.Book

	if err := config.DB.Where("isbn = ?", isbn).First(&book).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}

	var updatedBook models.Book
	if err := c.ShouldBindJSON(&updatedBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := bookValidator.Validate(updatedBook); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	updatedBook.ID = book.ID
	updatedBook.ISBN = isbn

	if err := config.DB.Save(&updatedBook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating book"})
		return
	}

	response := models.BookResponse{
		ID:        updatedBook.ID,
		CreatedAt: updatedBook.CreatedAt.Format(time.RFC3339),
		UpdatedAt: updatedBook.UpdatedAt.Format(time.RFC3339),
		ISBN:      updatedBook.ISBN,
		Title:     updatedBook.Title,
		Author:    updatedBook.Author,
		Publisher: updatedBook.Publisher,
		Price:     updatedBook.Price,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Delete a book
// @Description Delete a book by ISBN
// @Tags books
// @Accept json
// @Produce json
// @Param isbn path string true "Book ISBN"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books/{isbn} [delete]
func DeleteBook(c *gin.Context) {
	isbn := c.Param("isbn")
	var book models.Book

	if err := config.DB.Where("isbn = ?", isbn).First(&book).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}

	if err := config.DB.Delete(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}