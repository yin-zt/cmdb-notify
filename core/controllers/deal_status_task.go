package controllers

import (
	"fmt"
	"github.com/yin-zt/cmdb-notify/core/models"
)

// DealRelationTask 用于处理消息订阅推送关于关系字段变更的数据
func (f OperationFieldService) DealHostStatus(tStatus <-chan *models.OperateHostStatus) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			OpeLog.Errorf("DealRelationTask 捕获到panic异常，recover恢复回来了，【err】为：%s", err)
		}
	}()
}
