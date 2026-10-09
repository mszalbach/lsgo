// Package main coordinates the other packages and starts the web server.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	config, err := parseFlags(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(2)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("address", config.addr),
		slog.String("folder", config.folder),
		slog.String("baseURL", config.baseURL),
	)

	slog.SetDefault(logger)
	err = run(config)
	if err != nil {
		os.Exit(1)
	}
}

func run(config *config) error {
	root, err := filesystem.NewRoot(config.folder)
	if err != nil {
		return fmt.Errorf("failed to open root folder: %w", err)
	}
	defer root.Close()

	renderer, err := web.NewHTMLRenderer(config.baseURL, assets.Templates, "html/base.tmpl")
	if err != nil {
		return fmt.Errorf("failed to create HTML renderer: %w", err)
	}

	webServer := web.NewRouter(config.baseURL, renderer, root, config.maxInlineFileSize, config.logSampleRate)

	server := http.Server{
		Addr:              config.addr,
		Handler:           webServer.Routes(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	cancelCtx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	stopContext, stop := signal.NotifyContext(cancelCtx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		slog.Info("Serving folder")
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			cancel(err)
		}
	}()

	<-stopContext.Done()
	slog.Info("Shutting down server ...")
	timeoutCtx, timeoutFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutFunc()

	stopErr := server.Shutdown(timeoutCtx)
	if stopErr != nil {
		return fmt.Errorf("failed to shutdown server: %w", stopErr)
	}

	err = context.Cause(stopContext)
	if !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server stopped with error: %w", err)
	}
	return nil
}
