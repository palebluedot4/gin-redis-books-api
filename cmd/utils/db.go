package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"gin-redis-books-api/cmd/model"
)

var (
	log = logrus.New()
	Ctx = context.Background()
	Rdb *redis.Client
)

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)
}

func InitRedis() error {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := Rdb.Ping(Ctx).Err(); err != nil {
		log.Errorf("failed to connect to Redis: %v", err)
		return err
	}
	log.Info("connected to Redis successfully")

	if err := storeBooksInRedis(); err != nil {
		log.Errorf("failed to store books in Redis: %v", err)
		return err
	}

	return nil
}

func storeBooksInRedis() error {
	file, err := os.Open("utils/books.txt")
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		book, err := parseBookFromLine(line)
		if err != nil {
			return fmt.Errorf("error parsing book from line: %v", err)
		}

		bookJSON, err := json.Marshal(book)
		if err != nil {
			return fmt.Errorf("error marshaling book to JSON: %v", err)
		}

		if err := Rdb.Set(Ctx, book.ISBN, bookJSON, 0).Err(); err != nil {
			return fmt.Errorf("error setting book in Redis: %v", err)
		}

	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %v", err)
	}

	return nil
}

func parseBookFromLine(line string) (*model.Book, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid number of fields in line: %v", line)
	}

	floatPrice, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing price: %v", err)
	}
	intStock, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing stock: %v", err)
	}

	return &model.Book{
		ISBN:   fields[0],
		Title:  fields[1],
		Author: fields[2],
		Price:  floatPrice,
		Stock:  intStock,
	}, nil
}

func GetBooksFromRedis() ([]*model.Book, error) {
	var books []*model.Book

	keys, err := Rdb.Keys(Ctx, "*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys from Redis: %v", err)
	}

	for _, key := range keys {
		val, err := Rdb.Get(Ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get book from Redis: %v", err)
		}

		var book *model.Book
		err = json.Unmarshal([]byte(val), &book)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal book JSON: %v", err)
		}

		books = append(books, book)
	}

	return books, nil
}

func FindBookByISBNFromRedis(isbn string) (*model.Book, error) {
	val, err := Rdb.Get(Ctx, isbn).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get book from Redis: %v", err)
	}

	var book *model.Book
	err = json.Unmarshal([]byte(val), &book)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal book JSON: %v", err)
	}

	return book, nil
}
