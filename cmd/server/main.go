package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sleep2death/nac"
)

func main() {
	// set to release mode
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	nac.Init(".config.toml")

	// run the server
	err := r.Run()
	if err != nil {
		panic(err)
	}
}
