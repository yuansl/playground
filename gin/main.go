package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
)

func main() {
	go func() {
		http.ListenAndServe(":6061", nil)
	}()

	gin.SetMode(gin.ReleaseMode)
	ginx := gin.New()
	ginx.Use(gin.Recovery())

	ginx.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello, World 👋!")
	})
	println(ginx.Run(":3001"))
}
