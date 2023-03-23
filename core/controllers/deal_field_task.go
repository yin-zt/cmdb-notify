package controllers

import (
	"fmt"
	"github.com/yin-zt/cmdb-notify/core/cmdb"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"github.com/yin-zt/cmdb-notify/core/models"
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

			DiffData := ftask.ChangeData
			if ftask.Pflag {
				objField = "P_"
			}

			result := f.FindNeedSearchFields(proSearchFields)
			objSearch := map[string]string{"instanceId": instanceId}
			postData := map[string]interface{}{"page_size": 100, "page": 1}
			postData["fields"] = result
			postData["query"] = objSearch
			findObj, err := cmdb.Easy.GetAllInstance(objId, postData, 1)
			if !err {
				OpeLog.Error("fail to search obj")
			}
			if len(findObj) != 1 {
				OpeLog.Errorf("can not find out this model: %s, instanceID: %s", objId, instanceId)
			}
			targetCmdbData := findObj[0]
			if val, ok := config.ModelStatusMap[objId]; ok {
				if targetCmdbData[val] != "online" {
					fmt.Println("offline")
				} else {
					finalData := f.AnalyFieldData(objId, targetCmdbData, proSearchFields)
					fmt.Println(finalData)
					fmt.Println(f.CheckIpPort(proSearchFields, DiffData, finalData))
				}
				fmt.Println("fffffffffffffffffffffff")
			} else {
				fmt.Println("has not status")
			}
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
	return &finalData
}

// AnalyFieldData 分析从cmdb获取到的数据，并返回适合上报cmdb接口的数据
func (f OperationFieldService) AnalyFieldData(model string, data map[string]interface{}, fdata map[string]string) map[string]interface{} {
	var retData = map[string]interface{}{}
	for key, val := range fdata {
		count := strings.Count(key, ".")
		switch count {
		case 0:
			findVal := data[key]
			findRealVal, _ := findVal.(string)
			f.MakeKeyVal(val, findRealVal, retData)
		case 1:
			firstKey := strings.Split(key, ".")[0]
			secondKey := strings.Split(key, ".")[1]
			relateData := data[firstKey]
			storeVal := ""
			for _, item := range relateData.([]interface{}) {
				fmt.Println(item)
				fmt.Printf("%T", item)
				itemVal := item.(map[string]interface{})
				if realVal, ok := itemVal[secondKey]; !ok {
					response = "返回的关联数据中没有这个键的值"
					panic(response)
				} else {
					storeVal = storeVal + ";" + realVal.(string)
				}
			}
			storeVal = strings.Trim(storeVal, ";")
			f.MakeKeyVal(val, storeVal, retData)
		}
	}
	f.MakePfieldVal(retData, fdata, data)
	if model == "HOST" {
		retData["exporterName"] = retData["ip"].(string) + "-" + "9100"
		retData["exporterPort"] = 9100
	} else {
		retData["exporterName"] = fmt.Sprintf("%s-%s", retData["ip"], retData["exporterPort"])
	}
	return retData
}

// MakeKeyVal 根据字典映射值中是否包含"."进行特定处理
func (f OperationFieldService) MakeKeyVal(key, addVal string, data map[string]interface{}) {
	if addVal == "" {
		return
	}
	if strings.Count(key, ".") == 1 {
		firstKey := strings.Split(key, ".")[0]
		secondKey := strings.Split(key, ".")[1]
		if _, ok := data[firstKey]; !ok {
			data[firstKey] = map[string]string{secondKey: addVal}
		} else {
			tempData := data[firstKey]
			dictData, ok := tempData.(map[string]string)
			if ok {
				if _, ok := dictData[secondKey]; ok {
					dictData[secondKey] = dictData[secondKey] + ";" + addVal
				} else {
					dictData[secondKey] = addVal
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
			tempVal := stringVal + ";" + addVal
			data[key] = strings.Trim(tempVal, ";")
		} else {
			data[key] = addVal

		}
	}
}

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
func (f *OperationFieldService) CheckIpPort(fdata map[string]string, changeData interface{}, wholeVal map[string]interface{}) string {
	var (
		portFlag, ipFlag          = "", ""
		portBool, ipBool          bool
		portStr, ipStr, targetStr string
	)
	for objIndex, objField := range fdata {
		if objField == "exporterPort" {
			portFlag = objIndex
		}
		if objField == "ip" {
			ipFlag = objIndex
		}
	}
	fmt.Println(changeData)
	fmt.Printf("%T", changeData)
	changeRealVal, ok := changeData.(map[string]interface{})
	if !ok {
		response = "修改的字段新旧值并不是map类型"
		panic(response)
	}
	if val, ok := changeRealVal[portFlag]; ok {
		portBool = true
		portStr = val.(string)
	}
	if val, ok := changeRealVal[ipFlag]; ok {
		ipBool = true
		ipStr = val.(string)
	}
	if ipBool && portBool {
		targetStr = ipStr + "_" + portStr
	} else if ipBool {
		portStr = wholeVal["exporterPort"].(string)
		targetStr = ipStr + "_" + portStr
	} else if portBool {
		ipStr = wholeVal["ip"].(string)
		targetStr = ipStr + "_" + portStr
	} else {
		targetStr = ""
	}
	return targetStr
}
