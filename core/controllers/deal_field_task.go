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
			OpeLog.Errorf("DealFieldTask 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			OpeLog.Flush()
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
				statusStr, ok := targetCmdbData[val].(string)
				if !ok {
					OpeLog.Errorf("此模型: %s, exporterState 字段非 字符串，实例为：%s", objId, instanceId)
					continue
				}
				if statusStr != "online" {
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
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			OpeLog.Errorf("AnalyFieldData 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			OpeLog.Flush()
		}
	}()
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
				itemVal, okl := item.(map[string]interface{})
				if !okl {
					OpeLog.Errorf("%v 竟然不是map[string]interface{} 格式", item)
					continue
				}
				if realVal, ok := itemVal[secondKey]; !ok {
					response = "返回的关联数据中没有这个键的值"
					continue
				} else {
					realStr, ok := realVal.(string)
					if !ok {
						OpeLog.Errorf("AnalyFieldData 分析: %v 竟然不是字符串", realVal)
						continue
					}
					storeVal = storeVal + ";" + realStr
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
				itemVal, ok := item.(map[string]interface{})
				if !ok {
					OpeLog.Errorf("AnalyFieldData is not map[string]interface{} wtih %v", item)
					continue
				}
				if secondLevelData, ok := itemVal[secondKey]; !ok {
					response = "返回的第二层关联数据中没有这个键的值"
					continue
				} else {
					for _, secondLevelVal := range secondLevelData.([]interface{}) {
						thirdLevelData, ok := secondLevelVal.(map[string]interface{})
						if !ok {
							OpeLog.Errorf("secondLevelVal is not map[string]interface{} with value: %v", secondLevelVal)
							continue
						}
						if thirdLevelVal, ok := thirdLevelData[thirdKey]; !ok {
							response = "返回的第三层关联数据中没有这个键的值"
							continue
						} else {
							thridLevelStr, ok := thirdLevelVal.(string)
							if !ok {
								OpeLog.Errorf("thirdLevelVal 的值: %v 竟然不是字符串", thirdLevelVal)
								continue
							}
							storeVal = storeVal + ";" + thridLevelStr
						}
					}
				}
			}
			f.MakeKeyVal(val, strings.Trim(storeVal, ";"), retData)
		}
	}
	f.MakePfieldVal(retData, fdata, data)
	if model == "HOST" {
		ipStr, ok := retData["ip"].(string)
		if !ok {
			OpeLog.Errorf("retData[ip]的值: %v 竟然不是字符串", retData)
			return finalRetData
		}
		retData["exporterName"] = ipStr + "-" + "9100"
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
				portStr, ok := portItem.(string)
				if !ok {
					OpeLog.Errorf("portValues列表的值中竟然有不是字符串的存在，:%v", portValues)
					continue
				}
				mTemp := make(map[string]interface{})
				gocopy.Copy(&mTemp, &retData)
				mTemp["exporterName"] = fmt.Sprintf("%s-%s", retData["ip"], portStr)
				mTemp["exporterPort"] = portStr
				mTemp["exporterType"] = strings.ToLower(model) + "-exporter"
				retData["serviceName"] = strings.ToUpper(model)
				finalRetData = append(finalRetData, mTemp)
			}
		}

	}
	return finalRetData
}

// MakeKeyVal 根据字典映射值中是否包含"."进行特定处理
// key为bigMap中的值
func (f OperationFieldService) MakeKeyVal(key string, addVal interface{}, data map[string]interface{}) {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			OpeLog.Errorf("MakeKeyVal 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			OpeLog.Flush()
		}
	}()
	if addVal == "" {
		return
	}
	if strings.Count(key, ".") == 1 {
		addValStr, ok := addVal.(string)
		if !ok {
			OpeLog.Errorf("addVal的值: %v 竟然不是字符串", addVal)
			return
		}
		firstKey := strings.Split(key, ".")[0]
		secondKey := strings.Split(key, ".")[1]
		if _, ok := data[firstKey]; !ok {
			data[firstKey] = map[string]string{secondKey: addValStr}
		} else {
			tempData := data[firstKey]
			dictData, ok := tempData.(map[string]string)
			if ok {
				if _, ok := dictData[secondKey]; ok {
					dictData[secondKey] = dictData[secondKey] + ";" + addValStr
				} else {
					dictData[secondKey] = addValStr
				}
				data[firstKey] = dictData
			}
		}
	} else {
		if singleField, ok := data[key]; ok {
			stringVal, ok := singleField.(string)
			if !ok {
				OpeLog.Errorf("singleField is not a string field with value: %v", singleField)
			} else {
				addValStr, ok := addVal.(string)
				if !ok {
					OpeLog.Errorf("this single key addVal is not string :%v", addVal)
				} else {
					tempVal := stringVal + ";" + addValStr
					data[key] = strings.Trim(tempVal, ";")
				}
			}
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
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			OpeLog.Errorf("MakePfieldVal 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			OpeLog.Flush()
		}
	}()
	for _, value := range fdata {
		if strings.HasPrefix(value, "P_") {
			dataVal, ok := cmdbData[value].(string)
			if !ok {
				OpeLog.Errorf("MakePfieldVal fdata字段非字符串, %v", cmdbData[value])
				continue
			}
			if dataVal == "" {
				OpeLog.Infof("dataVal 值为空")
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
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			OpeLog.Errorf("CheckIpPort 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			OpeLog.Flush()
		}
	}()
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
			oldNewPort, ok := portTempData.(map[string]interface{})
			if !ok {
				OpeLog.Errorf("portTempData value is not map[string]interface{}, with value:%v", portTempData)
				return nil
			}
			switch portVals := oldNewPort["old"].(type) {
			case []interface{}:
				for _, oneport := range portVals {
					onePortStr, ok := oneport.(string)
					if !ok {
						OpeLog.Errorf("端口列表元素非字符串, 值为: %v", oneport)
						continue
					}
					portList = append(portList, onePortStr)
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
			ipTempMap, ok := ipTempData.(map[string]interface{})
			if !ok {
				OpeLog.Errorf("ipTempData value is not map[string]interface{}, :%v", ipTempData)
				return nil
			}
			ipStrVal, ok := ipTempMap["old"].(string)
			if !ok {
				OpeLog.Errorf("ipTempMap value is not string, :%v", ipTempMap["old"])
				return nil
			}
			ipStr = ipStrVal
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
		ipStr, ok := wholeVal[0]["ip"].(string)
		if !ok {
			OpeLog.Errorf("wholeVal ip value is not string, with %v", wholeVal[0]["ip"])
			return nil
		}
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
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			OpeLog.Errorf("SearchExporterIdByName 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			OpeLog.Flush()
		}
	}()
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
