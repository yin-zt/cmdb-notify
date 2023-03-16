package controllers

import (
	"fmt"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"github.com/yin-zt/cmdb-notify/core/models"
	"github.com/yin-zt/cmdb-notify/utils"
	"net/http"
	"strings"
)

var (

	//	// 操作日志channel
	//	//OperateLogChan = make(chan *models.OperateLog, 30)
	//
	// OperateFieldChan 字段修改任务
	OperateFieldChan = make(chan *models.OperateField, 100)
	//
	//	// OperateRelationChan 关系修改任务
	//	OperateRelationChan = make(chan *models.OperateRelation, 100)

	C1 = Common{}
)

func ChangedObj(w http.ResponseWriter, r *http.Request) {
	Obj := &models.Cobj{}
	utils.ParseBody(r, Obj)
	topic := Obj.Topic
	model := strings.Split(topic, ".")
	objModel := model[len(model)-1]
	if val, ok := config.BigMap[objModel]; ok {
		if strings.Index(topic, "instance.modify.") != -1 {
			res := Obj.Data.ExtInfo.ChangeFields
			for _, value := range res {
				if _, ok := val[value]; ok {
					diff := Obj.Data.ExtInfo.DiffData
					nValue, oValue := C1.FindModifyVal(diff, value)
					task := models.OperateField{
						Model:      objModel,
						Field:      value,
						TargetId:   Obj.Data.TargetId,
						ChangeData: models.Diff{Old: oValue, New: nValue},
					}
					fmt.Println(task)
					_ = task
				} else {
					fmt.Println("nothing")
					continue
				}
			}
		}
		fmt.Println("hello")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		strr := []byte("hello world")
		w.Write(strr)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		strr := []byte("do not worry")
		w.Write(strr)
	}
}
