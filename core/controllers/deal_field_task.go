package controllers

import (
	"fmt"
	"github.com/yin-zt/cmdb-notify/core/models"
	"time"
)

var (
	OperationFieldTask = &OperationFieldService{}
)

type OperationFieldService struct {
}

func (f OperationFieldService) DealFieldTask(fic <-chan *models.OperateField) {
	timer := time.NewTimer(50 * time.Second)
	defer timer.Stop()
	for {
		fmt.Println("111111111111111111111")
		select {
		case ftask := <-fic:
			fmt.Println(ftask)
		case <-timer.C: //5s同步一次
			fmt.Println("okr")
			timer.Reset(50 * time.Second)
		}
	}
}
