package models

type Cobj struct {
	System string `json:"system"`
	Topic  string `json:"topic"`
	Data   Data   `json:"data"`
}

type Data struct {
	Event          string  `json:"event"`
	EventId        string  `json:"event_id"`
	ExtInfo        ExtInfo `json:"ext_info"`
	Memo           string  `json:"memo"`
	TargetCategory string  `json:"target_category"`
	TargetId       string  `json:"target_id"`
	TargetName     string  `json:"target_name"`
}

type ExtInfo struct {
	ChangeFields []string    `json:"_change_fields"`
	DiffData     interface{} `json:"diff_data"`
	InstanceId   string      `json:"instance_id"`
	InstanceName string      `json:"instance_name"`
	ObjectId     string      `json:"object_id"`
}

type DiffData struct {
}
