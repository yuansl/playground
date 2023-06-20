package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	app := gin.New()

	app.Use(gin.Recovery())

	app.GET("/v2/admin/domains/:domain", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"uid": 999999, "hub": "live260"})
	})
	app.Run(":9000")
}
