package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/zhangpetergo/chat/chat/app/sdk/mux"
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var build = "develop"

func main() {
	// 初始化日志
	logger.InitLogger()

	ctx := context.Background()
	if err := run(ctx); err != nil {
		logger.Log.Errorw("startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS
	logger.Log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration
	cfg := struct {
		Version struct {
			Build string
			Desc  string
		}
		Web struct {
			ReadTimeout        time.Duration
			WriteTimeout       time.Duration
			IdleTimeout        time.Duration
			ShutdownTimeout    time.Duration
			APIHost            string
			CORSAllowedOrigins []string
		}
	}{
		Version: struct {
			Build string
			Desc  string
		}{build, "sales service"},
	}

	// 设置配置默认值
	// web
	viper.SetDefault("Web.ReadTimeout", "5s")
	viper.SetDefault("Web.WriteTimeout", "10s")
	viper.SetDefault("Web.IdleTimeout", "120s")
	viper.SetDefault("Web.ShutdownTimeout", "20s")
	viper.SetDefault("Web.APIHost", "0.0.0.0:9000")
	viper.SetDefault("Web.CORSAllowedOrigins", "*")

	// 设置配置文件路径和名称
	configPath := "./zarf/config"
	configName := "config"

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// 检查文件是否存在
	if _, err := os.Stat(configPath + "/" + configName + ".yaml"); err == nil {
		// 文件存在，读取配置文件
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		// 文件不存在，使用默认配置
	} else {
		// 其他错误
		return err
	}

	// 读取配置
	err := viper.Unmarshal(&cfg)
	if err != nil {
		// 解析配置失败
		return err
	}

	// -------------------------------------------------------------------------
	// App Starting

	logger.Log.Infow("starting service", "version", cfg.Version.Build)
	defer logger.Log.Info("shutdown complete")

	logger.Log.Infow("startup", "config", cfg)
	logger.BuildInfo()

	// -------------------------------------------------------------------------
	// Start API Service

	logger.Log.Infow("startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	webAPI := mux.WebAPI()

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      webAPI,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		//ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Log.Infow("startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer logger.Log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
