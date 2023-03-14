package routes

import (
	"github.com/gorilla/mux"
	"github.com/yin-zt/cmdb-notify/core/controllers"
)

var RegisterServerRouters = func(router *mux.Router) {
	router.HandleFunc("/cmdb5/notify/exporter/", controllers.ChangedObj).Methods("POST")
}
