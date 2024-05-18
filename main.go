package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Book struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var Db *sql.DB

func setupDatabase() {
	connectionString := fmt.Sprintf(
		"user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Fatal(err)
	}

	Db = db

	createUsersTable()
	createBooksTable()

}

func createUsersTable() {
	_, err := Db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(200) )
    `)

	if err != nil {
		fmt.Println(err)
	}
}

func createBooksTable() {
	_, err := Db.Exec(`
        CREATE TABLE IF NOT EXISTS books (
            id SERIAL PRIMARY KEY,
            title TEXT,
			author TEXT
		)
    `)

	if err != nil {
		fmt.Println(err)
	}
}

func handleAddUsers(context *gin.Context) {
	var user User

	err := context.BindJSON(&user)

	if err != nil {
		fmt.Println(err)
		context.JSON(400, gin.H{
			"message": "Bad request",
		})
		return
	}

	query := `
        INSERT INTO users (name)
        VALUES ($1)
        RETURNING id, name`

	err = Db.QueryRow(query, user.Name).Scan(
		&user.Id,
		&user.Name,
	)

	if err != nil {
		fmt.Println(err)
		context.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	context.JSON(201, user)
}

func handleAddBooks(context *gin.Context) {
	var book Book

	err := context.BindJSON(&book)

	if err != nil {
		fmt.Println(err)
		context.JSON(400, gin.H{
			"message": "Bad request",
		})
		return
	}

	query := `
        INSERT INTO books (title, author)
        VALUES ($1, $2)
        RETURNING id`

	err = Db.QueryRow(query, book.Title, book.Author).Scan(
		&book.Id,
	)

	if err != nil {
		fmt.Println(err)
		context.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	context.JSON(201, book)
}

func handleFetchBooks(context *gin.Context) {
	rows, err := Db.Query("SELECT * FROM books")

	if err != nil {

		log.Fatal(err)

		context.JSON(500, gin.H{
			"message": "Something went wrong",
		})
	}

	defer rows.Close()

	var books []Book

	for rows.Next() {

		var book Book

		if err := rows.Scan(&book.Id, &book.Title, &book.Author); err != nil {

			log.Fatal(err)

			context.JSON(500, gin.H{
				"message": "Something went wrong",
			})
		}

		books = append(books, book)
	}

	if err = rows.Err(); err != nil {

		log.Fatal(err)

		context.JSON(500, gin.H{
			"message": "Something went wrong",
		})
	}

	context.JSON(200, books)
}

func handleUpdateBooks(context *gin.Context) {
	context.JSON(200, gin.H{
		"message": "Updated successfully",
	})
}

func handleDeleteBooks(context *gin.Context) {

	query := `
      DELETE FROM books WHERE id=$1;`

	_, err := Db.Query(query, context.Param("id"))

	if err != nil {
		fmt.Println(err)
		context.JSON(500, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	context.Status(204)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	fmt.Println("Welcome to book app!")

	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
	}

	setupDatabase()

	if err != nil {
		fmt.Println(err)
	}

	router := gin.Default()

	router.Use(CORSMiddleware())

	router.POST("/users", handleAddUsers)
	router.GET("/books", handleFetchBooks)
	router.POST("/books", handleAddBooks)
	router.PUT("/books", handleUpdateBooks)
	router.DELETE("/books/:id", handleDeleteBooks)
	router.Run()

	defer Db.Close()
}
