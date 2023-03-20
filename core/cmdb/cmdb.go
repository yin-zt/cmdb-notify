package cmdb

import (
	"fmt"
)

func Test() {
	var easyApi = NewEasyapi("xxxx", "xxxxxx", "xxxxxxxxxxxxxxxxxxx")
	/*
	   // POST request
	   fmt.Println("-------POST EXAMPLE-------")
	   fieldInPostData := map[string]int{"ip": 1, "hostname": 1, "APP.name": 1}
	   postData :=  map[string]interface{} {"page_size":30, "page":1}
	   postData["fields"] = fieldInPostData
	   fmt.Println("[DATA]: ",postData)
	   ret ,isSuccess := easyApi.SendRequest("/cmdb_resource/object/HOST/instance/_search", "POST", postData)
	   fmt.Println("[Result]", ret, isSuccess)

	   // GET request
	   fmt.Println("-------GET EXAMPLE-------")
	   getData :=  map[string]interface{} {}
	   fmt.Println("[DATA]: ",getData)
	   ret ,isSuccess = easyApi.SendRequest("/cmdb_resource/object/APP/instance/5c1064fd42f06", "GET", getData)
	   fmt.Println("[Result]", ret, isSuccess)*/

	fmt.Println("-------GetAllInstance-------")
	fieldInPostData := map[string]int{"instanceId": 1, "name": 1}
	objSearch := map[string]string{"instanceId": "5d791458db54d"}
	postData := map[string]interface{}{"page_size": 3000, "page": 1}
	postData["fields"] = fieldInPostData
	postData["query"] = objSearch
	ret, isSuccess := easyApi.GetAllInstance("BUSINESSLEVEL", postData, 3)
	fmt.Println("[Result]", ret, isSuccess)
	retMap, isSuccess := easyApi.ChangeListToMap(ret, []string{"name"})
	fmt.Println("[Result Map]", retMap, isSuccess)
}
