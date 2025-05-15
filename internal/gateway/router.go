package gateway

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter() *mux.Router {
	return mux.NewRouter()
}

func SetupRouter(router *mux.Router) {
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	router.HandleFunc("/upload", Upload).Methods(http.MethodPost)
	router.HandleFunc("/download", Download).Methods(http.MethodGet)
	router.HandleFunc("/directory", CreateDirectory).Methods(http.MethodPost)
	router.HandleFunc("/delete", Delete).Methods(http.MethodDelete)
	router.HandleFunc("/list", List).Methods(http.MethodGet)
	router.HandleFunc("/move", Move).Methods(http.MethodPut)
	router.HandleFunc("/update", Update).Methods(http.MethodPut)
	router.HandleFunc("/copy", Copy).Methods(http.MethodPut)
}

func ListenAndServe(addr string, router *mux.Router) error {
	return http.ListenAndServe(addr, router)
}
