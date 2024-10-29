package main

import (
	"context"
	"cyberedge/pkg/api"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/service"
	"cyberedge/pkg/tasks"
	"cyberedge/pkg/utils"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 初始化日志系统
	if err := utils.InitializeLogging("logs"); err != nil {
		panic(err)
	}
	defer utils.StopLogging()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := utils.ConnectToMongoDB("mongodb://localhost:27017")
	if err != nil {
		return
	}
	defer utils.DisconnectMongoDB(client)

	db := client.Database("cyberedgeDB")
	taskDAO := dao.NewTaskDAO(db.Collection("tasks"))

	asynqClient, err := utils.InitAsynqClient("localhost:6379")
	if err != nil {
		return
	}
	defer asynqClient.Close()

	taskService := service.NewTaskService(taskDAO, asynqClient)

	asynqServer, err := utils.InitAsynqServer("localhost:6379")
	if err != nil {
		return
	}

	taskHandler := tasks.NewTaskHandler()
	pingTask := tasks.NewPingTask(taskDAO)
	taskHandler.RegisterHandler(tasks.TaskTypePing, pingTask.Handle)

	go func() {
		if err := asynqServer.Run(taskHandler); err != nil {
			panic(err)
		}
	}()

	router := api.NewRouter(
		nil,
		nil,
		taskService,
		"your-jwt-secret",
		"your-session-secret",
	)
	engine := router.SetupRouter()

	srv, err := utils.StartHTTPServer(":8081", engine)
	if err != nil {
		panic(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	utils.ShutdownHTTPServer(srv, 5*time.Second)
	asynqServer.Shutdown()
}
