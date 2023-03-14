package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func ParseBody(r *http.Request, x interface{}) {
	fmt.Println(r.Body)
	if body, err := ioutil.ReadAll(r.Body); err == nil {
		fmt.Println(body)
		fmt.Println(string(body))
		if err := json.Unmarshal(body, x); err != nil {
			return
		}
	}
}
