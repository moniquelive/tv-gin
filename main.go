package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

var r = gin.Default()

func init() {
	r.GET("/ping", pingHandler)
}

func main() {
	if err := r.Run(); err != nil {
		log.Fatalln("r.Run:", err)
	}
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
