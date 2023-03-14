package controllers

import (
	"github.com/yin-zt/cmdb-notify/core/models"
	"github.com/yin-zt/cmdb-notify/utils"
	"net/http"
)

func ChangedObj(w http.ResponseWriter, r *http.Request) {
	Obj := models.Cobj{}
	utils.ParseBody(r, Obj)
}
