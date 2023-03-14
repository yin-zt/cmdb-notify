package server

import (
	"github.com/gorilla/mux"
	"github.com/yin-zt/cmdb-notify/core/routes"
	"log"
	"net/http"
)

func Start() {
	r := mux.NewRouter()
	routes.RegisterServerRouters(r)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":9999", r))
}
