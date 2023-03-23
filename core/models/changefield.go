package models

type OperateField struct {
	Model      string                 `json:"model"`
	Field      string                 `json:"field"`
	TargetId   string                 `json:"target_id"`
	Pflag      bool                   `json:"pflag"`
	ChangeData map[string]interface{} `json:"diff_data"`
}
