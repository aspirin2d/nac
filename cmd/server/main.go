package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sleep2death/nac"
)

func main() {
	// set to release mode
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// setup routes
	n := nac.FromConfig(".config.toml")

	r.Handle("GET", "/ping", n.Ping)

	// run the server
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
