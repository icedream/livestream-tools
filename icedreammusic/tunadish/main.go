package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/icedream/livestream-tools/icedreammusic/tuna"
)

const addr = "localhost:1608"

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Baggage", "Content-Length", "Content-Type", "Access-Control-Allow-Headers", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	currentData := &tuna.TunaData{
		Status: tuna.Stopped,
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, *currentData)
	})

	r.POST("/", func(c *gin.Context) {
		newData := new(tuna.TunaRequest)

		var err error

		if err = c.BindJSON(newData); err == nil {
			if newData == nil {
				err = errors.New("invalid null request body")
			}
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		currentData = &newData.Data
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	log.Println("Listening on ", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
