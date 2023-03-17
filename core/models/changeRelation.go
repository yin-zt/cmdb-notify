package models

type OperateRelation struct {
	Model    string `json:"model"`
	Field    string `json:"field"`
	TargetId string `json:"target_id"`
	Flag     bool   `json:"checkPorNot"`
}
