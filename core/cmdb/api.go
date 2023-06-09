package cmdb

import (
	"fmt"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"github.com/yin-zt/cmdb-notify/core/loger"

	//"os"

	"bytes"
	"strconv"
	"strings"

	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"

	"encoding/hex"
	"encoding/json"

	"io/ioutil"
	"net/http"
	"net/url"

	"sort"
	"time"
)

var (
	respon       any
	Easy         = NewEasyapi(config.CmdbHost, config.CmdbAk, config.CmdbSk)
	cmdbOpeLoger = loger.GetLoggerOperate()
)

// EasyApi specifies a tool using EasyOps CMDB api
type Easyapi struct {
	cmdbAddr string
	ak       string
	sk       string
	header   map[string]string
}

// NewEasyapi create a new EasyApi
func NewEasyapi(cmdbAddr, ak, sk string) *Easyapi {
	header := map[string]string{"Host": "openapi.easyops-only.com", "Content-Type": "application/json;charset=UTF-8"}
	return &Easyapi{cmdbAddr, ak, sk, header}
}

// ChangeListToMap Change the Map list([] map[string]interface{}) to the Map which key is selected as keys
// srcList: the map list
// keys : the result map's key, if more than one, use | to join
// the result is the map[string]interface{}
func (ez *Easyapi) ChangeListToMap(srcList []map[string]interface{}, keys []string) (map[string]interface{}, bool) {
	var isSuccess = false
	retMap := make(map[string]interface{})
	lenKeys := len(keys)
	if lenKeys == 0 {
		keys = append(keys, "instanceId")
	}
	for _, ins := range srcList {
		fmt.Println(ins)
		var keyFlagList []string
		for _, key := range keys {
			keyFlagList = append(keyFlagList, ConvInterfaceToString(ins[key]))
		}
		fmt.Println(keyFlagList)
		keyFlagString := strings.Join(keyFlagList, "|")
		retMap[keyFlagString] = ins
	}
	return retMap, isSuccess
}

// GetAllInstance to get the all instances of the target object
// objectId: the CMDB resource object id
// the return is the list of map[string]interface{}
func (ez *Easyapi) GetAllInstance(objectId string, params map[string]interface{}, pagesize int) ([]map[string]interface{}, bool) {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			cmdbOpeLoger.Errorf("GetAllInstance 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			cmdbOpeLoger.Flush()
		}
	}()
	var isSuccess = false
	var result []map[string]interface{}
	if pagesize > 2000 {
		pagesize = 2000
	}
	params["page_size"] = pagesize
	params["page"] = 1
	var url = "/cmdbservice/object/" + objectId + "/instance/_search"
	reqRet, reqIsSuccess := ez.SendRequest(url, "POST", params)
	if reqIsSuccess {
		reqMap := ConvStringToMap(reqRet)
		reqData := reqMap["data"].(map[string]interface{})
		totalNum, err := strconv.Atoi(strconv.FormatInt(int64(reqData["total"].(float64)), 10))
		if err != nil {
			fmt.Println("[Fatal error] ", err.Error())
			return result, isSuccess
		}
		req, ok := reqData["list"].([]interface{})
		if !ok {
			return nil, false
		}
		for _, i := range req {
			result = append(result, i.(map[string]interface{}))
		}
		// need next page
		if totalNum > pagesize {
			pages := totalNum / pagesize
			for p := 2; p < pages+2; p++ {
				params["page"] = p
				reqRet, reqIsSuccess = ez.SendRequest(url, "POST", params)
				reqMap = ConvStringToMap(reqRet)
				req, ok := reqMap["data"].(map[string]interface{})["list"].([]interface{})
				if !ok {
					return nil, false
				}
				for _, i := range req {
					result = append(result, i.(map[string]interface{}))
				}
			}
		}
	} else {
		fmt.Println("[Request ERROR!!!]", reqRet)
		return result, isSuccess
	}

	return result, reqIsSuccess
}

// UpdateOrCreateObjs 更新或者创建模型实例
func (ez *Easyapi) UpdateOrCreateObjs(objectId string, key []string, postData []map[string]interface{}) {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			cmdbOpeLoger.Errorf("UpdateOrCreateObjs 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			cmdbOpeLoger.Flush()
		}
	}()
	var (
		params = map[string]interface{}{"keys": key, "datas": postData}
		url    = "/cmdbservice/object/" + objectId + "/instance/_import"
	)
	reqRet, reqIsSuccess := ez.SendRequest(url, "POST", params)
	fmt.Println(reqRet, reqIsSuccess)
	cmdbOpeLoger.Info(reqRet)
	cmdbOpeLoger.Info(reqIsSuccess)
}

// GetModelFieldsWithP 获取模型中的以P_开头的属性字段
func (ez *Easyapi) GetModelFieldsWithP(objectId string) ([]string, bool) {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			cmdbOpeLoger.Errorf("GetModelFieldsWithP 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			cmdbOpeLoger.Flush()
		}
	}()
	var (
		isSuccess bool
		result    []string
		//reqData map[string]interface{}
		param = map[string]interface{}{}
		url   = "/cmdb/object/attr/" + objectId
	)
	reqRet, reqIsSuccess := ez.SendRequest(url, "Get", param)
	if reqIsSuccess {
		reqMap := ConvStringToMap(reqRet)
		reqData, ok := reqMap["data"].([]interface{})
		if ok {
			for _, value := range reqData {
				perField := value.(map[string]interface{})
				filedVal := perField["id"]
				fieldStr := fmt.Sprintf("%s", filedVal)
				if strings.HasPrefix(fieldStr, "P_") {
					result = append(result, fieldStr)
				}
			}
			isSuccess = true
		} else {
			cmdbOpeLoger.Errorf("获取此模型字段异常，返回的数据非列表, 模型为：%s", objectId)
		}
	}
	return result, isSuccess
}

// ArchiveObject 用于归档模型实例数据
func (ez *Easyapi) ArchiveObject(modelId, instanceId string) {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			cmdbOpeLoger.Errorf("ArchiveObject 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			cmdbOpeLoger.Flush()
		}
	}()

	archiveUrl := fmt.Sprintf("/cmdbservice/object/%s/instance_archive/%s", modelId, instanceId)
	data := make(map[string]interface{})
	reqRet, responResult := ez.SendRequest(archiveUrl, "POST", data)
	if !responResult {
		fmt.Println("fail to call cmdb5 archive api")
	} else {
		fmt.Println(reqRet)
	}
}

// SendRequest to EasyOps OpenApi
func (ez *Easyapi) SendRequest(reqUrl string, method string, params map[string]interface{}) (string, bool) {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			cmdbOpeLoger.Errorf("SendRequest 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			cmdbOpeLoger.Flush()
		}
	}()
	var isSuccess = false
	var ret = ""
	// timestamp
	var nowTS = strconv.FormatInt(time.Now().Unix(), 10)
	//nowTS = "1546957559"
	method = strings.ToUpper(method)
	var sign = ez.genSignature(reqUrl, method, params, nowTS)

	var requestUrl = "http://" + ez.cmdbAddr + reqUrl
	name := url.Values{"accesskey": {ez.ak}, "expires": {nowTS}, "signature": {sign}}
	param := name.Encode()
	requestUrl = fmt.Sprintf("%s?%s", requestUrl, param)
	//requestUrl = requestUrl + "?accesskey=" + ez.ak + "&expires=" + nowTS + "&signature=" + sign
	client := http.Client{}
	if method == "GET" || method == "DELETE" {
		if params != nil {
			for k, v := range params {
				requestUrl = requestUrl + "&" + k + "=" + ConvInterfaceToString(v)
			}
		}
		req, err := http.NewRequest(method, requestUrl, nil)
		if err != nil {
			fmt.Println("[Fatal error] ", err.Error())
			return ret, isSuccess
		}
		for k, v := range ez.header {
			req.Header.Set(k, v)
		}
		//  the Host header is promoted to the Request.Host field and removed from the Header map.
		req.Host = "openapi.easyops-only.com"
		response, err := client.Do(req)
		if err != nil {
			fmt.Println("[Fatal error] ", err.Error())
			return ret, isSuccess
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("[Fatal error] ", err.Error())
			return ret, isSuccess
		}
		isSuccess = true
		ret = string(body)

	} else if method == "POST" || method == "PUT" {
		bytesData, err := json.Marshal(params)
		if err != nil {
			fmt.Println("[Fatal error] ", err.Error())
			return ret, isSuccess
		}
		reader := bytes.NewReader(bytesData)

		req, err := http.NewRequest(method, requestUrl, reader)
		if err != nil {
			fmt.Println("[Fatal error] ", err.Error())
			return ret, isSuccess
		}
		for k, v := range ez.header {
			req.Header.Set(k, v)
		}
		//  the Host header is promoted to the Request.Host field and removed from the Header map.
		req.Host = "openapi.easyops-only.com"

		response, err := client.Do(req)
		if err != nil {
			fmt.Println("[Fatal error] ", err.Error())
			return ret, isSuccess
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("[Fatal error] ", err.Error())
			return ret, isSuccess
		}

		isSuccess = true
		ret = string(body)
	} else {
		respon = "Request method not known"
		panic(respon)
	}

	return ret, isSuccess
}

// hmacSHA1Encrypt encrypt the encryptText use encryptKey
func (ez *Easyapi) hmacSHA1Encrypt(encryptText, encryptKey string) string {
	key := []byte(encryptKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(encryptText))
	var str string = hex.EncodeToString(mac.Sum(nil))
	//fmt.Printf("[encrypt result] %v\n", str)
	return str
}

// genSignature generate the Signature using in easyosp api request
/*
Signature = HMAC-SHA1('SecretKey', UTF-8-Encoding-Of( StringToSign ) ) );
StringToSign = HTTP-Verb + "\n" +
               URL + "\n" +
               Parameters + "\n" +
               Content-Type + "\n" +
               Content-MD5 + "\n" +
               Date + "\n" +
               AccessKey;
*/
func (ez *Easyapi) genSignature(url string, method string, params map[string]interface{}, nowTS string) string {
	var resp any
	defer func() {
		if err := recover(); err != resp {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			cmdbOpeLoger.Errorf("genSignature 捕获到panic异常，recover并没有恢复，【err】为：%s", err)
			cmdbOpeLoger.Flush()
		}
	}()

	//fmt.Println(reflect.ValueOf(nowTS))
	var urlParams string = ""
	var bodyContent string = ""
	if params != nil {
		// method is GET or DELETE , params build  the url_params
		method = strings.ToUpper(method)
		if method == "GET" || method == "DELETE" {
			// sort the params
			keys := SortMapByKey(params)
			for _, k := range keys {
				urlParams = urlParams + k + ConvInterfaceToString(params[k])
			}
			// method is POST or PUT, params build the bodyContent
		} else if method == "POST" || method == "PUT" {
			jsonStr, err := json.Marshal(params)
			//fmt.Println(string(jsonStr))
			//fmt.Println(reflect.TypeOf(jsonStr))
			if err != nil {
				respon = err
				panic(respon)
			}
			md5Ctx := md5.New()
			md5Ctx.Write(jsonStr)
			cipherStr := md5Ctx.Sum(nil)
			bodyContent = hex.EncodeToString(cipherStr)
		} else {
			fmt.Println(method)
			respon = "Request method not known"
			panic(respon)
		}
	}

	// HTTP-Verb + "\n" +URL + "\n" +Parameters + "\n" +Content-Type + "\n" +Content-MD5 + "\n" +Date + "\n" +AccessKey;
	var str_sign = method + "\n" + url + "\n" + urlParams + "\n" + ez.header["Content-Type"] + "\n" + bodyContent + "\n" + nowTS + "\n" + ez.ak
	/*
	   fmt.Println("-------------------------------")
	   fmt.Println("before encrypt:\n"+str_sign)
	   fmt.Println("-------------------------------")*/
	var sign = ez.hmacSHA1Encrypt(str_sign, ez.sk)
	return sign
}

// SortMapByKey sort the key of map and return the sorted key slice
func SortMapByKey(OriMap map[string]interface{}) []string {
	keys := make([]string, len(OriMap))
	i := 0
	for k, _ := range OriMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// ConvInterfaceToString translate the interface{} to string
func ConvInterfaceToString(i interface{}) string {
	var ret string
	switch i.(type) {
	case string:
		ret = i.(string)
		break
	case int:
		ret = strconv.FormatInt(int64(i.(int)), 10)
		break
	default:
		fmt.Println("params type not supported", i)
		respon = "params type not supported"
		panic(respon)
	}
	return ret
}

// ConvStringToMap
func ConvStringToMap(reqRet string) map[string]interface{} {
	var info map[string]interface{}
	err := json.Unmarshal([]byte(reqRet), &info)
	if err != nil {
		fmt.Println("json Unmarshal error: ", err.Error())
		return info
	}
	return info
}
