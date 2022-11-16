package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
		colors = append(colors, color)
	}

	colorsBytes, _ := json.Marshal(colors)
	c.IndentedJSON(http.StatusOK, colorsBytes)
}

var Router *gin.Engine

func main() {
	fmt.Println(os.Getenv("DATABASE_URL"))

	r := gin.Default()
	r.GET("/", GETColors)
	r.Run()
}
