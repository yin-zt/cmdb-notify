package server

import (
	"context"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/gorilla/mux"
	"github.com/yin-zt/cmdb-notify/core/controllers"
	loger2 "github.com/yin-zt/cmdb-notify/core/loger"
	"github.com/yin-zt/cmdb-notify/core/routes"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func Start() {
	defer log.Flush()
	_loger := loger2.GetLoggerAcc()
	log.ReplaceLogger(_loger)
	r := mux.NewRouter()
	routes.RegisterServerRouters(r)
	http.Handle("/", r)
	for i := 0; i < 10; i++ {
		go controllers.OperationFieldTask.DealFieldTask(controllers.OperateFieldChan)
	}
	for i := 0; i < 10; i++ {
		go controllers.OperationFieldTask.DealRelationTask(controllers.OperateRelationChan)
	}
	//for i := 0; i < 3; i++ {
	//	go controllers.OperationFieldTask.DealHostStatus(controllers.OperateStatusChan)
	//}

	go func() {
		if err := http.ListenAndServe(":9999", r); err != nil && err != http.ErrServerClosed {
			log.Errorf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Error("server exiting")
	fmt.Println("test only")
}
