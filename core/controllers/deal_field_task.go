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

var (
	OperationFieldTask = &OperationFieldService{}
	response           any
)

type OperationFieldService struct {
}

// DealFieldTask 用于处理消息订阅推送关于字段变更的数据
func (f OperationFieldService) DealFieldTask(fic <-chan *models.OperateField) {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			OpeLog.Errorf("DealFieldTask 捕获到panic异常，recover恢复回来了，【err】为：%s", err)
		}
	}()
	//var needSearch map[string]int
	//var tempMap map[string]string
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case ftask := <-fic:
			objId := ftask.Model
			instanceId := ftask.TargetId
			objField := ftask.Field
			proSearchFields := config.BigMap[objId][objField]
			DiffData := ftask.ChangeData
			if ftask.Pflag {
				if pFieldResult, ok := cmdb.Easy.GetModelFieldsWithP(objId); ok {
					for _, pVal := range pFieldResult {
						proSearchFields[pVal] = fmt.Sprintf("customLabel.%s", pVal)
					}
				} else {
					OpeLog.Errorf("fail to search model P_ field with models:[%s]", objId)
					continue
				}
			}
			var fieldResult = map[string]int{}
			f.FindNeedSearchFields(proSearchFields, fieldResult)
			objSearch := map[string]string{"instanceId": instanceId}
			postData := map[string]interface{}{"page_size": 100, "page": 1}
			postData["fields"] = fieldResult
			postData["query"] = objSearch
			findObj, err := cmdb.Easy.GetAllInstance(objId, postData, 1)
			OpeLog.Infof("%v", findObj)
			fmt.Println(len(findObj))
			fmt.Println(findObj)
			if !err {
				OpeLog.Error("deal_field_task fail to search obj")
				continue
			}
			if len(findObj) != 1 {
				OpeLog.Errorf("can not find out this model: %s, instanceID: %s", objId, instanceId)
				continue
			}
			targetCmdbData := findObj[0]
			OpeLog.Infof("%v", targetCmdbData)
			fmt.Println(targetCmdbData)
			if val, ok := config.ModelStatusMap[objId]; ok {
				if targetCmdbData[val].(string) != "online" {
					OpeLog.Infof("This object status is not online, instanceID: %s; model: %s", instanceId, objId)
					continue
				} else {
					finalData := f.AnalyFieldData(objId, targetCmdbData, proSearchFields)
					fmt.Println(finalData)
					cmdb.Easy.UpdateOrCreateObjs("EXPORTER", []string{"exporterName"}, finalData)
					OpeLog.Infof("create or update one exporter instance with data: %s", finalData)
					//fmt.Println(finalData)
					if len(finalData) >= 1 {
						needArchiveExporter := f.CheckIpPort(proSearchFields, DiffData, finalData)
						OpeLog.Infof("need to archive exporter obj with data: %s", strings.Join(needArchiveExporter, ","))
						fmt.Println(needArchiveExporter)
						if len(needArchiveExporter) != 0 {
							for _, archiveOne := range needArchiveExporter {
								archiveExporterId := f.SearchExporterIdByName(archiveOne)
								if archiveExporterId != "" {
									cmdb.Easy.ArchiveObject("EXPORTER", archiveExporterId)
									OpeLog.Infof("success to archive obj with instanceID:[%s]", archiveExporterId)
								} else {
									OpeLog.Info("archive exporter id is null")
								}
							}
						}
					} else {
						OpeLog.Errorf("this instanceID: %s, has null postData", instanceId)
					}
				}
			} else {
				fmt.Println("has not status")
				OpeLog.Error("this change field task does not has exporterstatus field")
			}
		case <-timer.C: //5s同步一次
			fmt.Println("okr")
			timer.Reset(50 * time.Second)
		}
	}
}

// FindNeedSearchFields 将字段组合成适合调用cmdb接口的格式
func (f OperationFieldService) FindNeedSearchFields(retData map[string]string, finalData map[string]int) {
	for key, _ := range retData {
		finalData[key] = 1
	}
}

// AnalyFieldData 分析从cmdb获取到的数据，并返回适合上报cmdb接口的数据
func (f OperationFieldService) AnalyFieldData(model string, data map[string]interface{}, fdata map[string]string) []map[string]interface{} {
	var finalRetData []map[string]interface{}
	var retData = map[string]interface{}{}
	for key, val := range fdata {
		count := strings.Count(key, ".")
		switch count {
		case 0:
			findVal := data[key]
			switch findRealVal := findVal.(type) {
			case string:
				f.MakeKeyVal(val, findRealVal, retData)
			case []interface{}:
				f.MakeKeyVal(val, findRealVal, retData)
			}
		case 1:
			firstKey := strings.Split(key, ".")[0]
			secondKey := strings.Split(key, ".")[1]
			relateData := data[firstKey]
			storeVal := ""
			for _, item := range relateData.([]interface{}) {
				itemVal := item.(map[string]interface{})
				if realVal, ok := itemVal[secondKey]; !ok {
					response = "返回的关联数据中没有这个键的值"
					continue
				} else {
					storeVal = storeVal + ";" + realVal.(string)
				}
			}
			storeVal = strings.Trim(storeVal, ";")
			f.MakeKeyVal(val, storeVal, retData)
		case 2:
			firstKey := strings.Split(key, ".")[0]
			secondKey := strings.Split(key, ".")[1]
			thirdKey := strings.Split(key, ".")[2]
			firstLevelData := data[firstKey]
			storeVal := ""
			for _, item := range firstLevelData.([]interface{}) {
				itemVal := item.(map[string]interface{})
				if secondLevelData, ok := itemVal[secondKey]; !ok {
					response = "返回的第二层关联数据中没有这个键的值"
					continue
				} else {
					for _, secondLevelVal := range secondLevelData.([]interface{}) {
						thirdLevelData := secondLevelVal.(map[string]interface{})
						if thirdLevelVal, ok := thirdLevelData[thirdKey]; !ok {
							response = "返回的第三层关联数据中没有这个键的值"
							continue
						} else {
							storeVal = storeVal + ";" + thirdLevelVal.(string)
						}
					}
				}
			}
			f.MakeKeyVal(val, strings.Trim(storeVal, ";"), retData)
		}
	}
	f.MakePfieldVal(retData, fdata, data)
	if model == "HOST" {
		retData["exporterName"] = retData["ip"].(string) + "-" + "9100"
		retData["exporterPort"] = 9100
		retData["exporterType"] = "host" + "-exporter"
		retData["serviceName"] = strings.ToUpper(model)
		finalRetData = append(finalRetData, retData)
	} else {
		switch portValues := retData["exporterPort"].(type) {
		case string:
			retData["exporterName"] = fmt.Sprintf("%s-%s", retData["ip"], retData["exporterPort"])
			retData["exporterType"] = strings.ToLower(model) + "-exporter"
			retData["serviceName"] = strings.ToUpper(model)
			finalRetData = append(finalRetData, retData)
		case []interface{}:
			for _, portItem := range portValues {
				mTemp := make(map[string]interface{})
				gocopy.Copy(&mTemp, &retData)
				mTemp["exporterName"] = fmt.Sprintf("%s-%s", retData["ip"], portItem.(string))
				mTemp["exporterPort"] = portItem.(string)
				mTemp["exporterType"] = strings.ToLower(model) + "-exporter"
				retData["serviceName"] = strings.ToUpper(model)
				finalRetData = append(finalRetData, mTemp)
			}
		}

	}
	return finalRetData
}

// MakeKeyVal 根据字典映射值中是否包含"."进行特定处理
func (f OperationFieldService) MakeKeyVal(key string, addVal interface{}, data map[string]interface{}) {
	if addVal == "" {
		return
	}
	if strings.Count(key, ".") == 1 {
		firstKey := strings.Split(key, ".")[0]
		secondKey := strings.Split(key, ".")[1]
		if _, ok := data[firstKey]; !ok {
			data[firstKey] = map[string]string{secondKey: addVal.(string)}
		} else {
			tempData := data[firstKey]
			dictData, ok := tempData.(map[string]string)
			if ok {
				if _, ok := dictData[secondKey]; ok {
					dictData[secondKey] = dictData[secondKey] + ";" + addVal.(string)
				} else {
					dictData[secondKey] = addVal.(string)
				}
				data[firstKey] = dictData
			}
		}
	} else {
		if singleField, ok := data[key]; ok {
			stringVal, ok := singleField.(string)
			if !ok {
				response = "singleField is not a string field"
				panic(response)
			}
			tempVal := stringVal + ";" + addVal.(string)
			data[key] = strings.Trim(tempVal, ";")
		} else {
			data[key] = addVal

		}
	}
}

//func (f OperationFieldService) MakeKeyMultiVal(key, addVal []interface{}, data map[string]interface{}){
//	if len(addVal) == 0{
//		return
//	}
//
//}

func (f OperationFieldService) MakePfieldVal(data map[string]interface{}, fdata map[string]string, cmdbData map[string]interface{}) {
	for _, value := range fdata {
		if strings.HasPrefix(value, "P_") {
			dataVal := cmdbData[value].(string)
			if dataVal == "" {
				continue
			}
			if _, ok := data["customLabel"]; !ok {
				data["customLabel"] = map[string]string{value: dataVal}
			} else {
				tempData := data["customLabel"]
				dictData, ok := tempData.(map[string]string)
				if ok {
					if _, ok := dictData[value]; ok {
						dictData[value] = dictData[value] + ";" + dataVal
					} else {
						dictData[value] = dataVal
					}
					data["customLabel"] = dictData
				}
			}
		}
	}
}

// CheckIpPort 检查修改字段中是否包含ip和port字段
func (f *OperationFieldService) CheckIpPort(fdata map[string]string, changeData map[string]interface{}, wholeVal []map[string]interface{}) []string {
	var (
		portFlag, ipFlag              = "", ""
		portBool, ipBool              bool
		ipStr, targetStr              string
		portList, returnName, delName []string
	)
	for objIndex, objField := range fdata {
		if objField == "exporterPort" {
			portFlag = objIndex
		}
		if objField == "ip" {
			ipFlag = objIndex
		}
	}
	if portFlag != "" {
		portTempData := changeData[portFlag]
		if portTempData != nil {
			portBool = true
			oldNewPort := portTempData.(map[string]interface{})
			switch portVals := oldNewPort["old"].(type) {
			case []interface{}:
				for _, oneport := range portVals {
					portList = append(portList, oneport.(string))
				}
			case string:
				portList = append(portList, portVals)
			}
		}
	}
	if ipFlag != "" {
		ipTempData := changeData[ipFlag]
		if ipTempData != nil {
			ipBool = true
			ipTempMap := ipTempData.(map[string]interface{})
			ipStr = ipTempMap["old"].(string)
		}
	}
	if ipBool && portBool {
		for _, onePort := range portList {
			targetStr = ipStr + "-" + onePort
			returnName = append(returnName, targetStr)
		}
	} else if ipBool {
		for _, nowItem := range wholeVal {
			portInfo := nowItem["exporterPort"]
			switch port := portInfo.(type) {
			case string:
				returnName = append(returnName, fmt.Sprintf("%s-%s", ipStr, port))
			case int:
				returnName = append(returnName, fmt.Sprintf("%s-%d", ipStr, port))
			}
		}
	} else if portBool {
		if len(wholeVal) == 0 {
			return returnName
		}
		ipStr = wholeVal[0]["ip"].(string)
		for _, portVal := range portList {
			returnName = append(returnName, fmt.Sprintf("%s-%s", ipStr, portVal))
		}
	} else {
		OpeLog.Infof("do not delete ip or port field")
		return delName
	}
	for _, oneStr := range returnName {
		var begin bool
		for _, oneObj := range wholeVal {
			if val := oneObj["exporterName"]; val != nil {
				if val == oneStr {
					begin = true
					break
				}
			}
		}
		if !begin {
			delName = append(delName, oneStr)
		}
	}
	return delName
}

func (f *OperationFieldService) SearchExporterIdByName(exporterName string) string {
	fieldInPostData := map[string]int{"instanceId": 1, "name": 1}
	objSearch := map[string]string{"exporterName": exporterName}
	postData := map[string]interface{}{"page_size": 3, "page": 1}
	postData["fields"] = fieldInPostData
	postData["query"] = objSearch
	ret, isSuccess := cmdb.Easy.GetAllInstance("EXPORTER", postData, 100)
	if isSuccess {
		for _, oneObj := range ret {
			id := oneObj["instanceId"]
			return fmt.Sprintf("%s", id)
		}
	} else {
		OpeLog.Infof("can not find this name's 【%s】 instanceId", exporterName)
	}
	return ""
}
