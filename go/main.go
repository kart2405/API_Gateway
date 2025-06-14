package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Albums struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []Albums{
	{ID: "1", Title: "Extralu", Artist: "MB", Price: 100},
	{ID: "2", Title: "Aagadu", Artist: "", Price: 100},
}

func postalbums(c *gin.Context) {
	var newalbum Albums
	if err := c.BindJSON(&newalbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	albums = append(albums, newalbum)
	c.IndentedJSON(http.StatusCreated, newalbum)

}
func getalbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}
func albumsid(c *gin.Context) {
	id := c.Param("id")
	for _, album := range albums {
		if album.ID == id {
			c.IndentedJSON(http.StatusOK, album)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album with id " + id + " not found"})
}

func logMiddleware(c *gin.Context) {
	fmt.Println("log Middleware")
	c.Next()
}

func main() {
	r := gin.Default() //to create the instance of the gin
	r.Use(logMiddleware)
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World"})

	})
	r.GET("/home", func(p *gin.Context) {
		p.JSON(http.StatusOK, gin.H{"Bane": "Extralu"})
	})

	r.POST("/post", postalbums)
	r.GET("/albums", getalbums) //to get the albums
	r.GET("/album/:id", albumsid)
	r.Run(":8080")

}
