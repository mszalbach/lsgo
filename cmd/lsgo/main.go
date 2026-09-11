// Package main coordinates the other packages and starts the webserver
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

	"github.com/mszalbach/lsgo/internal/explorer"
	"github.com/mszalbach/lsgo/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := flag.String("addr", "localhost:8080", "Address to listen on. Default only listens on localhost.")
	folder := flag.String("folder", "./public", "Folder to serve.")
	flag.Parse()

	root, err := explorer.NewRoot(*folder)
	if err != nil {
		slog.Error("Could not open root folder", slog.String("folder", *folder), slog.Any("error", err))
		os.Exit(1)
	}

	webServer, err := web.NewServer(root)
	if err != nil {
		slog.Error("Could not create handler for web server", slog.Any("error", err))
		os.Exit(1)
	}

	server := http.Server{
		Addr:              *addr,
		Handler:           webServer.Router(),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
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
