package main

import (
	"log"

	"github.com/koan6gi/go-drive/internal/app/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
