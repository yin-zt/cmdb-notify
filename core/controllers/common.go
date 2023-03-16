package controllers

import "fmt"

type Common struct {
}

// FindModifyVal 从post数据提取目标字段前后变更数据
func (m *Common) FindModifyVal(Target map[string]interface{}, field string) (new, old string) {
	var newOld = Target[field]
	switch newOld.(type) {
	case map[string]interface{}:
		finalData := newOld.(map[string]interface{})
		old := fmt.Sprintf("%v", finalData["old"])
		new := fmt.Sprintf("%v", finalData["new"])
		return new, old
	}
	return "", ""
}
