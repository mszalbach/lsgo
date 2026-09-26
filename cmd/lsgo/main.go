// Package main coordinates the other packages and starts the web server.
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
	// needed for scratch images to have timezone information
	_ "time/tzdata"

	"github.com/mszalbach/lsgo/internal/assets"
	"github.com/mszalbach/lsgo/internal/filesystem"
	"github.com/mszalbach/lsgo/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := parseFlags(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(2)
	}

	root, err := filesystem.NewRoot(config.folder)
	if err != nil {
		slog.Error("Failed to open root folder", slog.String("folder", config.folder), slog.Any("error", err))
		os.Exit(1)
	}
	defer root.Close()

	renderer, err := web.NewHTMLRenderer(config.baseURL, assets.Templates, "html/base.tmpl")
	if err != nil {
		slog.Error("Failed to create handler for web server", slog.Any("error", err))
		panic(err)
	}

	webServer := web.NewRouter(root, renderer, config.maxInlineFileSize, config.logSampleRate)

	server := http.Server{
		Addr:              config.addr,
		Handler:           webServer.Routes(),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	stopContext, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		slog.Info("Serving folder", "address", config.addr, "folder", config.folder)
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start server", "error", err)
			panic(err)
		}
	}()

	<-stopContext.Done()
	slog.Info("Graceful shutdown")
	timeoutCtx, timeoutFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutFunc()

	stopErr := server.Shutdown(timeoutCtx)
	if stopErr != nil {
		slog.Warn("Failed to stop server", slog.Any("error", stopErr))
	}
}
