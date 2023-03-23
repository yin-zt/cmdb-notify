package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"github.com/yin-zt/cmdb-notify/core/loger"
	"github.com/yin-zt/cmdb-notify/core/models"
	"github.com/yin-zt/cmdb-notify/utils"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	//	// 操作日志channel
	OpeLog = loger.GetLoggerOperate()

	// OperateFieldChan 字段修改任务
	OperateFieldChan = make(chan *models.OperateField, 100)

	// OperateRelationChan 关系修改任务
	OperateRelationChan = make(chan *models.OperateRelation, 100)

	C1 = Common{}
)

// ChangedObj 用于处理cmdb消息订阅推送过来的数据，判断变更字段是否为监听字段；
// 若是，则解析请求体并将信息以任务形式传入channel中，由特定routine来处理
func ChangedObj(w http.ResponseWriter, r *http.Request) {
	defer OpeLog.Flush()
	TestObj := &models.AllModel{}
	bodyByte, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyByte))
	json.Unmarshal(bodyByte, TestObj)
	topic := TestObj.Topic
	model := strings.Split(topic, ".")
	changedModel := model[len(model)-1]
	if val, ok := config.BigMap[changedModel]; ok {
		if strings.Index(topic, "instance.modify.") != -1 {
			Obj := &models.Cobj{}
			utils.ParseBody(r, Obj)
			res := Obj.Data.ExtInfo.ChangeFields
			for _, value := range res {
				if _, ok := val[value]; ok {
					diff := Obj.Data.ExtInfo.DiffData
					//nValue, oValue := C1.FindModifyVal(diff, value)
					cTask := &models.OperateField{
						Model:    changedModel,
						Field:    value,
						TargetId: Obj.Data.TargetId,
						Pflag:    false,
						//ChangeData: models.Diff{Old: oValue, New: nValue},
						ChangeData: diff,
					}
					OperateFieldChan <- cTask
					OpeLog.Infof("success to send a field changed task to channel %v", &cTask)
				} else if strings.HasPrefix("P_", value) || config.SelfDefineField[fmt.Sprintf("%s_%s", changedModel, value)] {
					diff := Obj.Data.ExtInfo.DiffData
					cTask := &models.OperateField{
						Model:    changedModel,
						Field:    value,
						TargetId: Obj.Data.TargetId,
						Pflag:    true,
						//ChangeData: models.Diff{Old: "", New: ""},
						ChangeData: diff,
					}
					OperateFieldChan <- cTask
					OpeLog.Infof("success to send a field changed task to channel %v", &cTask)
				} else {
					OpeLog.Info("no match")
					OpeLog.Info(Obj.Topic, Obj.Data)
					continue
				}
			}
		} else if strings.Index(topic, "instance_relation.create.") != -1 {
			ChgObj := &models.RelObj{}
			utils.ParseBody(r, ChgObj)
			relateField := ChgObj.Data.ExtInfo.ChangedRel
			if _, ok := config.BigMap[relateField]; ok {
				rTask := &models.OperateRelation{
					Model:    changedModel,
					Field:    relateField,
					TargetId: ChgObj.Data.ExtInfo.TargetId,
				}
				OperateRelationChan <- rTask
				OpeLog.Infof("success to send a relation change task to channel %v", &rTask)
			} else {
				OpeLog.Info("no match")
				OpeLog.Info(ChgObj.Topic, ChgObj.Data)
			}
		} else {
			fmt.Println("do something")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		strr := []byte("hello world")
		w.Write(strr)
	} else {
		OpeLog.Infof("do not carry about this model: %v", changedModel)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		strr := []byte("i don't moniter this model")
		w.Write(strr)
	}
}
