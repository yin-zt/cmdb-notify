package models

type RelObj struct {
	System string `json:"system"`
	Topic  string `json:"topic"`
	Data   Data1  `json:"data"`
}

type Data1 struct {
	Event          string   `json:"event"`
	EventId        string   `json:"event_id"`
	ExtInfo        ExtInfo1 `json:"ext_info"`
	Memo           string   `json:"memo"`
	TargetCategory string   `json:"target_category"`
	TargetId       string   `json:"target_id"`
	TargetName     string   `json:"target_name"`
}

type ExtInfo1 struct {
	ChangedRel    string `json:"relation_side_id"`
	DstInstanceId string `json:"dst_instance_id"`
}
