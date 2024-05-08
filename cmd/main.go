package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"gin-redis-books-api/cmd/controller"
	"gin-redis-books-api/cmd/utils"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	if err := utils.InitRedis(); err != nil {
		log.Printf("failed to initialize Redis: %v", err)
	}

	router := gin.Default()
	router.GET("/books", controller.GetBooksHandler)
	router.GET("/books/:isbn", controller.GetBookByISBNHandler)
	router.POST("/books", controller.CreateBookHandler)
	router.DELETE("/books/:isbn", controller.DeleteBookHandler)
	router.PATCH("/books/:isbn", controller.UpdateBookHandler)

	log.Fatal(router.Run("localhost:8080"))
}
