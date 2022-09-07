package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Dsmit05/img-loader/internal/broker/stream"
	"github.com/Dsmit05/img-loader/internal/config"
	"github.com/Dsmit05/img-loader/internal/logger"
	"github.com/Dsmit05/img-loader/internal/service"
	"github.com/Dsmit05/img-loader/pkg/downloader"
	pool "github.com/Dsmit05/img-loader/pkg/worker-pool"
)

func main() {
	if err := logger.InitLogger(false, "logs.json"); err != nil {
		panic(err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("config.NewConfig() ", err)
	}

	consumerGRPC := stream.NewImageListener(cfg)
	go consumerGRPC.StartGRPC()
	go consumerGRPC.StartREST()

	pl, err := pool.NewDynamicWorkersPool(4, 50, true)
	if err != nil {
		logger.Fatal("pool.NewDynamicWorkersPool() ", err)
		return
	}

	loader, err := downloader.NewFileLoader(downloader.Config{CopyBufferSize: 1024, BasePath: "Users"})
	if err != nil {
		logger.Fatal("downloader.NewFileLoader() ", err)
		return
	}

	// run main app
	app, err := service.NewSaveLoader(pl, loader, consumerGRPC)
	if err != nil {
		logger.Fatal("service.NewSaveLoader() ", err)
		return
	}
	app.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	consumerGRPC.Stop()
	app.Stop()

}
