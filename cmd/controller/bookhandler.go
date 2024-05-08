package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-redis-books-api/cmd/model"
	"gin-redis-books-api/cmd/utils"
)

func GetBooksHandler(ctx *gin.Context) {
	books, err := utils.GetBooksFromRedis()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to get books from Redis"})
		return
	}

	ctx.IndentedJSON(http.StatusOK, books)
}

func FindBookByISBN(isbn string) *model.Book {
	book, err := utils.FindBookByISBNFromRedis(isbn)
	if err != nil {
		return nil
	}

	return book
}

func GetBookByISBNHandler(ctx *gin.Context) {
	isbn := ctx.Param("isbn")
	book := FindBookByISBN(isbn)
	if book == nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	ctx.IndentedJSON(http.StatusOK, book)
}

func CreateBookHandler(ctx *gin.Context) {
	var newBook *model.Book

	if err := ctx.BindJSON(&newBook); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid book data"})
		return
	}

	newBookJSON, err := json.Marshal(newBook)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal book to JSON"})
		return
	}

	if err := utils.Rdb.Set(utils.Ctx, newBook.ISBN, newBookJSON, 0).Err(); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to store book in Redis: %v", err)})
		return
	}

	ctx.IndentedJSON(http.StatusOK, newBook)
}

func DeleteBookHandler(ctx *gin.Context) {
	isbn := ctx.Param("isbn")

	if err := utils.Rdb.Del(utils.Ctx, isbn).Err(); err != nil {
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "book deleted"})
}

func UpdateBookHandler(ctx *gin.Context) {
	isbn := ctx.Param("isbn")
	book, err := utils.FindBookByISBNFromRedis(isbn)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to get book from Redis"})
		return
	}
	if book == nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "Book not found", "isbn": isbn})
		return
	}

	var updatedBook model.Book
	if err := ctx.BindJSON(&updatedBook); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid update data"})
		return
	}

	if updatedBook.ISBN != "" {
		book.ISBN = updatedBook.ISBN
	}
	if updatedBook.Title != "" {
		book.Title = updatedBook.Title
	}
	if updatedBook.Author != "" {
		book.Author = updatedBook.Author
	}
	if updatedBook.Price != 0 {
		book.Price = updatedBook.Price
	}
	if updatedBook.Stock != 0 {
		book.Stock = updatedBook.Stock
	}

	updatedBookJSON, err := json.Marshal(book)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal updated book to JSON"})
		return
	}

	if err := utils.Rdb.Set(utils.Ctx, book.ISBN, updatedBookJSON, 0).Err(); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update book in Redis: %v", err)})
		return
	}

	ctx.IndentedJSON(http.StatusOK, book)

}
