package models

type OperateField struct {
	Model      string `json:"model"`
	Field      string `json:"model"`
	TargetId   string `json:"target_id"`
	ChangeData Diff   `json:"diff_data"`
}

type Diff struct {
	Old string `json:"old"`
	New string `json:"new"`
}
