package models

type OperateHostStatus struct {
	ID        string `json:"instanceId"`
	NewStatus string `json:"new_status"`
}
