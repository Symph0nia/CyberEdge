package utils

import (
	"context"
	"cyberedge/pkg/logging"
	"net/http"
	"time"
)

// StartHTTPServer 启动 HTTP 服务器
func StartHTTPServer(addr string, handler http.Handler) (*http.Server, error) {
	logging.Info("尝试启动 HTTP 服务器，地址: %s", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("HTTP 服务器启动失败: %v", err)
			panic(err)
		}
	}()

	logging.Info("HTTP 服务器成功启动，地址: %s", addr)
	return srv, nil
}

// ShutdownHTTPServer 关闭 HTTP 服务器
func ShutdownHTTPServer(srv *http.Server, timeout time.Duration) error {
	logging.Info("尝试关闭 HTTP 服务器，超时时间: %v", timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logging.Error("HTTP 服务器关闭失败: %v", err)
		return err
	}

	logging.Info("HTTP 服务器成功关闭")
	return nil
}
