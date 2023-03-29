package models

type OperateHostStatus struct {
	ID        string `json:"instanceId"`
	IP        string `json:"ip"`
	NewStatus string `json:"new_status"`
}
