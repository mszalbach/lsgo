package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mszalbach/lsgo/internal/backend"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := flag.String("addr", "localhost:8080", "Address to listen on. Default only listens on localhost.")
	folder := flag.String("folder", "./public", "Folder to serve.")
	flag.Parse()

	server := http.Server{
		Addr:    *addr,
		Handler: backend.Router(),
	}

	stopContext, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		slog.Info("Serving folder", "address", *addr, "folder", *folder)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start server", "error", err)
			panic(err)
		}
	}()

	<-stopContext.Done()
	slog.Info("Graceful shutdown")
	timeoutCtx, timeoutFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutFunc()

	if stopErr := server.Shutdown(timeoutCtx); stopErr != nil {
		slog.Warn("server stop failed", slog.Any("error", stopErr))
	}
}
