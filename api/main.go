package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "templates/index.tmpl", gin.H{
			// TODO: Populate list of active rooms.
		})
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Request.URL.Path = "/static/favicon.ico"
		router.HandleContext(c)
	})

	// Start the server.
	err := router.Run(":5000")
	if err != nil {
		log.Fatalf("error from main: %+v", err)
	}
}
