package controllers

import (
	"fmt"
	"github.com/yin-zt/cmdb-notify/core/cmdb"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"github.com/yin-zt/cmdb-notify/core/models"
	"time"
)

// DealRelationTask 用于处理消息订阅推送关于关系字段变更的数据
func (f OperationFieldService) DealRelationTask(ric <-chan *models.OperateRelation) {
	//var needSearch map[string]int
	//var tempMap map[string]string
	timer := time.NewTimer(50 * time.Second)
	defer timer.Stop()
	for {
		select {
		case rTask := <-ric:
			objId := rTask.Model
			instanceId := rTask.TargetId
			objField := rTask.Field
			proSearchFields := config.BigMap[objId][objField]
			fieldResult := f.FindNeedSearchFields(proSearchFields)
			objSearch := map[string]string{"instanceId": instanceId}
			postData := map[string]interface{}{"page_size": 100, "page": 1}
			postData["fields"] = fieldResult
			postData["query"] = objSearch
			findObj, err := cmdb.Easy.GetAllInstance(objId, postData, 1)
			if !err {
				OpeLog.Error("fail to search obj")
			}
			if len(findObj) == 0 {
				OpeLog.Errorf("can not find out this model: %s, instanceID: %s", objId, instanceId)
				continue
			}
			if len(findObj) == 0 {
				return
			}
			targetCmdbData := findObj[0]
			if statusField, ok := config.ModelStatusMap[objId]; ok {
				if targetCmdbData[statusField].(string) != "online" {
					OpeLog.Infof("This object status is not online, instanceID: %s; model: %s", instanceId, objId)
					fmt.Println("This object status is not online, instanceID")
					return
				}
			} else {

			}
		case <-timer.C: //5s同步一次
			fmt.Println("relationship okr")
			timer.Reset(50 * time.Second)
		}
	}
}

// AnalyRelationData 处理关系字段变更的情况，分为两种：第一种是变更模型自身存在ExporterStatus字段；另一种是自身不存在ExporterStatus字段
// model 为变更模型； cmdbData 为需要查询的数据   rfields为字段映射
//func (f OperationFieldService) AnalyRelationData(model string, cmdbData map[string]interface{}, rfields map[string]string) {
//	for leftKey, rightKey := range rfields {
//		count := strings.Count(leftKey, ".")
//		switch count {
//		case 0:
//
//		}
//	}
//}

// AnalyOwnRelationData 处理模型(模型需要自身有exporterStatus字段)关联关系变更的数据
// 特征：cmdb返回的数据中，会包含三种层级，类似(REDIS_REDISINFO.REDISINFO_ENV.env_en); 但返回exporter数据多少只由当前变更实例的exporterPort决定
//func (f OperationFieldService) AnalyOwnRelationData(model string, cmdbData map[string]interface{}, rfields map[string]string) {
//	for leftKey, rightKey := range rfields {
//		count := strings.Count(leftKey)
//		switch count {
//		case 0
//		}
//	}
//}
