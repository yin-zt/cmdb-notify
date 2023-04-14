package models

type CreateObj struct {
	System string     `json:"system"`
	Topic  string     `json:"topic"`
	Data   CreateData `json:"data"`
}

type CreateData struct {
	Event          string        `json:"event"`
	EventId        string        `json:"event_id"`
	ExtInfo        CreateExtInfo `json:"ext_info"`
	Memo           string        `json:"memo"`
	TargetCategory string        `json:"target_category"`
	TargetId       string        `json:"target_id"` // 实例id
	TargetName     string        `json:"target_name"`
}

type CreateExtInfo struct {
	InstanceId string `json:"instance_id"`
}
