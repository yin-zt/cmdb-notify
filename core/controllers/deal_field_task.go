package controllers

import (
	"fmt"
	"github.com/yin-zt/cmdb-notify/core/cmdb"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"github.com/yin-zt/cmdb-notify/core/models"
	"time"
)

var (
	OperationFieldTask = &OperationFieldService{}
)

type OperationFieldService struct {
}

// DealFieldTask 用于处理消息订阅推送关于字段变更的数据
func (f OperationFieldService) DealFieldTask(fic <-chan *models.OperateField) {
	//var needSearch map[string]int
	//var tempMap map[string]string
	timer := time.NewTimer(50 * time.Second)
	defer timer.Stop()
	for {
		select {
		case ftask := <-fic:
			objId := ftask.Model
			instanceId := ftask.TargetId
			objField := ftask.Field
			proSearchFields := config.BigMap[objId][objField]
			result := f.FindNeedSearchFields(proSearchFields)
			objSearch := map[string]string{"instanceId": instanceId}
			postData := map[string]interface{}{"page_size": 100, "page": 1}
			postData["fields"] = result
			postData["query"] = objSearch
			findObj, err := cmdb.Easy.GetAllInstance(objId, postData, 1)
			if !err {
				OpeLog.Error("fail to search obj")
			}
			fmt.Println(findObj)
		case <-timer.C: //5s同步一次
			fmt.Println("okr")
			timer.Reset(50 * time.Second)
		}
	}
}

// FindNeedSearchFields 将字段组合成适合调用cmdb接口的格式
func (f OperationFieldService) FindNeedSearchFields(retData map[string]string) *map[string]int {
	var finalData = map[string]int{}
	for key, _ := range retData {
		finalData[key] = 1
	}
	fmt.Println(finalData)
	return &finalData
}
