package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Color struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	ColorHex string `json:"colorhex"`
}

func OpenConnection() *sql.DB {
	var port, err = strconv.Atoi(os.Getenv("PGPORT"))
	fmt.Println("THIS IS PORT", port)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		os.Getenv("PGHOST"), port, os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"))

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected")
	return db
}

func GETColors(c *gin.Context) {
	db := OpenConnection()

	rows, err := db.Query(`SELECT * FROM colors`)
	if err != nil {
		log.Fatal()
	}
	var colors []Color

	for rows.Next() {
		var color Color
		rows.Scan(&color.Id, &color.Title, &color.ColorHex)
		fmt.Println(color)
		colors = append(colors, color)
	}

	c.IndentedJSON(http.StatusOK, colors)
}

func POSTColors(c *gin.Context) {
	db := OpenConnection()

	var color Color
	if err := c.BindJSON(&color); err != nil {
		return
	}

	insertQuery := `INSERT INTO colors (title, colorhex) VALUES ($1, $2)`
	_, err := db.Exec(insertQuery, color.Title, color.ColorHex)

	// Returns proper message based on whether duplicates are found
	if err != nil {
		if strings.HasPrefix(err.Error(), `pq: duplicate key`) {
			resp := make(map[string]string)
			if strings.HasSuffix(err.Error(), `"unique_title"`) {
				resp["message"] = "Duplicate color name"
			} else {
				resp["message"] = "Duplicate color code"
			}
			c.JSON(http.StatusBadRequest, resp)
			return
		} else {
			c.Writer.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
	defer db.Close()
}

// I'm using a post method since it was easier to do it this way
// + I was using the 'http' library in the desktop version to handle the endpoints
// and afaik no dynamic routing exists in that one
func DELETEColors(c *gin.Context) {
	db := OpenConnection()

	type colorId struct {
		Id int64 `json:"id"`
	}

	var id colorId
	if err := c.BindJSON(&id); err != nil {
		return
	}

	insertQuery := `DELETE FROM colors WHERE id=$1`
	_, err := db.Exec(insertQuery, id.Id)
	if err != nil {

		c.Writer.WriteHeader(http.StatusBadRequest)
	} else {
		c.Writer.WriteHeader(http.StatusOK)
	}
	defer db.Close()
}

var Router *gin.Engine

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/", GETColors)
	r.POST("/insert", POSTColors)
	r.POST("/delete", DELETEColors)
	r.Run()
}
