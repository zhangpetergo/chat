package main

import (
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	// 初始化日志
	logger.InitLogger()
	if err := run(); err != nil {
		logger.Log.Errorw("startup", "err", err)
		os.Exit(1)
	}
}

func run() error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS
	logger.Log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// App start
	logger.Log.Infow("startup", "status", "started")
	defer logger.Log.Infow("startup", "status", "shuttingdown")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown

	return nil
}
