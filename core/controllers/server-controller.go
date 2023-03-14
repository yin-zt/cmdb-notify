package controllers

import (
	"fmt"
	"github.com/yin-zt/cmdb-notify/core/models"
	"github.com/yin-zt/cmdb-notify/utils"
	"net/http"
)

var (
	// 操作日志channel
	//OperateLogChan = make(chan *models.OperateLog, 30)

	// OperateFieldChan 字段修改任务
	OperateFieldChan = make(chan *models.OperateField, 100)

	// OperateRelationChan 关系修改任务
	OperateRelationChan = make(chan *models.OperateRelation, 100)
)

func ChangedObj(w http.ResponseWriter, r *http.Request) {
	Obj := &models.Cobj{}
	utils.ParseBody(r, Obj)
	fmt.Println(Obj.System)
	fmt.Println(Obj.Data)
	fmt.Println(Obj.Topic)
	fmt.Println(Obj.Data.ExtInfo.DiffData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	strr := []byte("hello world")
	w.Write(strr)
}
