package controllers

import (
	"fmt"
	"github.com/yin-zt/cmdb-notify/core/cmdb"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"github.com/yin-zt/cmdb-notify/core/models"
	"github.com/young2j/gocopy"
	"strings"
	"time"
)

// DealRelationTask 用于处理消息订阅推送关于关系字段变更的数据
func (f OperationFieldService) DealRelationTask(ric <-chan *models.OperateRelation) {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover并没有恢复回来")
			OpeLog.Errorf("DealRelationTask 捕获到panic异常，recover并没有恢复回来了，【err】为：%s", err)
		}
	}()
	//var needSearch map[string]int
	//var tempMap map[string]string
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case rTask := <-ric:
			objId := rTask.Model
			instanceId := rTask.TargetId
			objField := rTask.Field
			proSearchFields := config.BigMap[objId][objField]
			if rTask.Flag {
				if pFieldResult, ok := cmdb.Easy.GetModelFieldsWithP(objId); ok {
					for _, pVal := range pFieldResult {
						proSearchFields[pVal] = fmt.Sprintf("customLabel.%s", pVal)
					}
				} else {
					OpeLog.Errorf("deal relation task can not find out this model fields info with ID: %v", objId)
					continue
				}
			}
			var fieldResult = map[string]int{}
			f.FindNeedSearchFields(proSearchFields, fieldResult)
			objSearch := map[string]string{"instanceId": instanceId}
			postData := map[string]interface{}{"page_size": 100, "page": 1}
			postData["fields"] = fieldResult
			postData["query"] = objSearch
			findObj, ok := cmdb.Easy.GetAllInstance(objId, postData, 1)
			if !ok {
				OpeLog.Error("deal_relation_task fail to search obj")
				continue
			}
			if len(findObj) == 0 {
				OpeLog.Errorf("can not find out this model: %s, instanceID: %s", objId, instanceId)
				continue
			}
			targetCmdbData := findObj[0]
			OpeLog.Infof("deal with data: %v", targetCmdbData)
			if statusField, ok := config.ModelStatusMap[objId]; ok {
				if targetCmdbData[statusField].(string) != "online" {
					OpeLog.Infof("This object status is not online, instanceID: %s; model: %s", instanceId, objId)
					continue
				} else {
					finalData := f.AnalyFieldData(objId, targetCmdbData, proSearchFields)
					cmdb.Easy.UpdateOrCreateObjs("EXPORTER", []string{"exporterName"}, finalData)
					OpeLog.Infof("deal_relation_task post data with %v", finalData)
					fmt.Println(finalData)
				}
			} else {
				finalData := f.AnalyOtherRelationData(objId, targetCmdbData, proSearchFields)
				cmdb.Easy.UpdateOrCreateObjs("EXPORTER", []string{"exporterName"}, finalData)
				OpeLog.Infof("%v", finalData)
				fmt.Println(finalData)
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

//AnalyOwnRelationData 处理模型(模型需要自身有exporterStatus字段)关联关系变更的数据
//特征：返回exporter数据多少只由当前变更实例的exporterPort决定
//func (f OperationFieldService) AnalyOwnRelationData(model string, cmdbData map[string]interface{}, rfields map[string]string) {
//	var finalRetData []map[string]interface{}
//	var retData = map[string]interface{}{}
//	for leftKey, rightKey := range rfields {
//		count := strings.Count(leftKey, ".")
//		switch count {
//		case 0:
//			if realValue, ok := cmdbData[leftKey]; !ok{
//				OpeLog.Errorf("do not find out about this [field: %s] in cmdb return [data: %s]", leftKey, cmdbData)
//			}else{
//				f.MakeKeyVal(rightKey, realValue,retData)
//			}
//		case 1:
//			firstKey := strings.Split(leftKey, ".")[0]
//			secondKey := strings.Split(leftKey, ".")[1]
//		}
//	}
//}

func (f OperationFieldService) AnalyOtherRelationData(model string, cmdbData map[string]interface{}, rfields map[string]string) []map[string]interface{} {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover并没有恢复回来")
			OpeLog.Errorf("AnalyOtherRelationData 捕获到panic异常，recover并没有恢复回来了，【err】为：%s", err)
		}
	}()
	publicKeyVal := make(map[string]interface{})
	hasSearchOrNot := make(map[string]bool)
	var finalRetData []map[string]interface{}
	var NotRelationObjs []string

	if notRelationObjs, ok := config.NotDirectionMapping[model]; !ok {
		return nil
	} else {
		NotRelationObjs = notRelationObjs
	}
	for leftKey, rightKey := range rfields {
		count := strings.Count(leftKey, ".")
		var relationFieldBool bool
		switch count {
		case 0:
			if realValue, ok := cmdbData[leftKey]; !ok {
				OpeLog.Errorf("do not find out about this [field: %s] in cmdb return [data: %s]", leftKey, cmdbData)
			} else {
				f.MakeKeyVal(rightKey, realValue, publicKeyVal)
			}
		case 1:
			firstKey := strings.Split(leftKey, ".")[0]
			secondKey := strings.Split(leftKey, ".")[1]
			for _, notDirectItem := range NotRelationObjs {
				if notDirectItem == firstKey {
					relationFieldBool = true
				}
			}
			if hasSearchOrNot[firstKey] {
				continue
			}
			if firstLevelData, ok := cmdbData[firstKey]; !ok {
				OpeLog.Errorf("do not find out about this [field: %s] in cmdb return [data: %s]", firstKey, cmdbData)
				return nil
			} else {
				if relationFieldBool {
					hasSearchOrNot[firstKey] = true
					needSearchFields := f.findAllNeedSearchField(rfields, firstKey)
					for _, oneItem := range firstLevelData.([]interface{}) {
						fmt.Println("test log")
						if dataMap, ok := oneItem.(map[string]interface{}); !ok {
							OpeLog.Errorf("二层关联字段在cmdb返回数据中并不是字典格式存在, data: %v", oneItem)
							continue
						} else {
							var oneResult = make(map[string]interface{})
							for onekey, onevalue := range needSearchFields {
								keyValue := dataMap[onekey]
								f.MakeKeyVal(onevalue, keyValue, oneResult)
							}
							if oneResult["state"] != "online" {
								continue
							}
							if model == "HOST" {
								oneResult["exporterName"] = oneResult["ip"].(string) + "-" + "9100"
								oneResult["exporterPort"] = 9100
								oneResult["exporterType"] = "host" + "-exporter"
								finalRetData = append(finalRetData, oneResult)
							} else {
								switch portValues := oneResult["exporterPort"].(type) {
								case string:
									oneResult["exporterName"] = fmt.Sprintf("%s-%s", oneResult["ip"], oneResult["exporterPort"])
									finalRetData = append(finalRetData, oneResult)
								case []interface{}:
									for _, portItem := range portValues {
										mTemp := make(map[string]interface{})
										gocopy.Copy(&mTemp, &oneResult)
										mTemp["exporterName"] = fmt.Sprintf("%s-%s", oneResult["ip"], portItem.(string))
										mTemp["exporterPort"] = portItem.(string)
										finalRetData = append(finalRetData, mTemp)
									}
								}
							}
						}
					}
				} else {
					for _, oneItem := range firstLevelData.([]interface{}) {
						if dataMap, ok := oneItem.(map[string]interface{}); !ok {
							OpeLog.Errorf("二层关联字段在cmdb返回数据中并不是字典格式存在, data: %v", oneItem)
							continue
						} else {
							if secondVal, ok := dataMap[secondKey]; !ok {
								OpeLog.Errorf("can not find out sample relation field value field:%s", leftKey)
							} else {
								f.MakeKeyVal(rightKey, secondVal, publicKeyVal)
							}
						}
					}
				}
			}
			for _, oneReturnObj := range finalRetData {
				for kk, vv := range publicKeyVal {
					oneReturnObj[kk] = vv
				}
			}
		}
	}
	return finalRetData
}

func (f OperationFieldService) findAllNeedSearchField(targetMap map[string]string, searchField string) map[string]string {
	returnMapFields := make(map[string]string)
	for key, val := range targetMap {
		switch count := strings.Count(key, "."); count {
		case 0:
			continue
		case 1:
			if strings.HasPrefix(key, searchField) {
				secondKey := strings.Split(key, ".")[1]
				returnMapFields[secondKey] = val
			}
		case 2:
			response = "do not think about this situation"
			panic(response)
		}
	}
	return returnMapFields
}
