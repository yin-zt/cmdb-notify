package controllers

import (
	"fmt"
	"github.com/yin-zt/cmdb-notify/core/models"
	"github.com/yin-zt/cmdb-notify/utils"
	"net/http"
)

func ChangedObj(w http.ResponseWriter, r *http.Request) {
	Obj := &models.Cobj{}
	utils.ParseBody(r, Obj)
	fmt.Println(&Obj)
	fmt.Println(Obj.System)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	strr := []byte("hello world")
	w.Write(strr)
}
