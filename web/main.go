package main

import (
	"log"
	"net/http"

	"github.com/bbeck/twitch-plays-crosswords/web/assets"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Serve everything under /web/static as a static asset.
	router.Use(assets.ServeStatic())

	// Configure everything under /web/templates/ as a template file.
	router.HTMLRender = assets.TemplateRender()

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
