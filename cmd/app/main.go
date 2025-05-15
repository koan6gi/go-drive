package main

import (
	"log"

	_ "github.com/koan6gi/go-drive/docs"

	"github.com/koan6gi/go-drive/internal/app/app"
)

// @title File Storage API
// @version 1.0
// @description API for file operations
// @host localhost:8080
// @BasePath /
// @schemes http
// @openapi 3.0.0
func main() {
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
