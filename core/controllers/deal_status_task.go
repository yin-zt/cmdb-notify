package controllers

import (
	"fmt"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"github.com/yin-zt/cmdb-notify/core/models"
	"time"
)

var ()

// DealRelationTask 用于处理消息订阅推送关于关系字段变更的数据
func (f OperationFieldService) DealHostStatus(tStatus <-chan *models.OperateHostStatus) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("捕获到了panic 产生的异常： ", err)
			fmt.Println("捕获到panic的异常了，recover恢复回来")
			OpeLog.Errorf("DealRelationTask 捕获到panic异常，recover恢复回来了，【err】为：%s", err)
		}
	}()
	timer := time.NewTimer(50 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C: //5s同步一次
			fmt.Println("cHostStatus okr")
			timer.Reset(50 * time.Second)
		case statusTask := <-tStatus:
			ip := statusTask.IP
			statusId := statusTask.NewStatus
			newStatus := config.StatusMap[statusId]

		}
	}

}
