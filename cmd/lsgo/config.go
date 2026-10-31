package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"path"
	"strconv"
	"strings"
)

type config struct {
	addr              string
	folder            string
	baseURL           string
	maxInlineFileSize int64
	logSampleRate     float64
}

func parseFlags(args []string, output io.Writer) (*config, error) {
	fs := flag.NewFlagSet("lsgo", flag.ContinueOnError)
	fs.SetOutput(output)
	var config config
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

	sampleRate := &boundedInt{
		Value: 0.05,
		Min:   0,
		Max:   1,
	}
	fs.Var(sampleRate, "log-sample-rate", "Sampling rate for successful, non-error logs (0.0 = 0%, 1.0 = 100%).")

	err := fs.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	config.logSampleRate = sampleRate.Value
	config.normalizeBaseURL()

	return &config, nil
}

func (c *config) normalizeBaseURL() {
	cleaned := path.Clean(c.baseURL)
	if cleaned == "." || cleaned == "/" {
		c.baseURL = ""
		return
	}
	c.baseURL = "/" + strings.Trim(cleaned, "/")
}

type boundedInt struct {
	Value float64
	Min   float64
	Max   float64
}

func (b *boundedInt) String() string {
	return strconv.FormatFloat(b.Value, 'f', -1, 64)
}

func (b *boundedInt) Set(s string) error {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("value %s is not a number (must be between %g and %g)", s, b.Min, b.Max)
	}
	if math.IsNaN(val) || val < b.Min || val > b.Max {
		return fmt.Errorf("value %g is out of range (must be between %g and %g)", val, b.Min, b.Max)
	}

	b.Value = val
	return nil
}
