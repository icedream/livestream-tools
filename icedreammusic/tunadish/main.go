package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TunaData struct {
	CoverURL string   `json:"cover_url"`
	Title    string   `json:"title"`
	Artists  []string `json:"artists"`
	Label    string   `json:"label"`
	Status   string   `json:"status"`
	Progress uint64   `json:"progress"`
	Duration uint64   `json:"duration"`
}

const addr = "localhost:1608"

func main() {
	r := gin.Default()

	currentData := &TunaData{
		Status: "stopped",
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, currentData)
	})

	r.POST("/", func(c *gin.Context) {
		newData := new(TunaData)

		if err := c.Bind(newData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		currentData = newData
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	log.Println("Listening on ", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
