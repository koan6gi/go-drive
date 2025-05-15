package app

import (
	"github.com/koan6gi/go-drive/internal/gateway"
	"github.com/koan6gi/go-drive/internal/repository"
)

const (
	addr = "localhost:8080"
)

func Run() error {
	var err error
	repository.FileStorage, err = repository.NewFileStorage()
	if err != nil {
		return err
	}

	router := gateway.NewRouter()
	gateway.SetupRouter(router)

	return gateway.ListenAndServe(addr, router)
}
