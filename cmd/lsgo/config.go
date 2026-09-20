package main

import (
	"flag"
	"fmt"
	"io"
	"path"
	"strings"
)

type Config struct {
	addr              string
	folder            string
	baseURL           string
	maxInlineFileSize int64
}

func parseFlags(args []string, output io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("lsgo", flag.ContinueOnError)
	fs.SetOutput(output)
	var config Config
	fs.StringVar(&config.addr, "addr", "localhost:8080", "Address to listen on. Default only listens on localhost.")
	fs.StringVar(&config.folder, "folder", "./public", "Folder to serve.")
	fs.StringVar(
		&config.baseURL,
		"base-url",
		"/",
		"Path webserver is served under. Required when running behind a reverse proxy with a path prefix.",
	)
	fs.Int64Var(
		&config.maxInlineFileSize,
		"max-inline-file-size",
		1_048_576,
		"Maximum file size to display inline in bytes; larger files are download-only.",
	)

	err := fs.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("could not parse config %w", err)
	}
	config.normalizeBaseURL()

	return &config, nil
}

func (c *Config) normalizeBaseURL() {
	cleaned := path.Clean(c.baseURL)
	if cleaned == "." || cleaned == "/" {
		c.baseURL = "/"
		return
	}
	c.baseURL = "/" + strings.Trim(cleaned, "/") + "/"
}
