package web

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

var safeInlineMediaTypes = map[string]bool{
	"text/plain":       true,
	"text/x-log":       true,
	"text/markdown":    true,
	"image/jpeg":       true,
	"image/png":        true,
	"image/gif":        true,
	"image/webp":       true,
	"audio/mpeg":       true,
	"audio/wav":        true,
	"audio/ogg":        true,
	"video/mp4":        true,
	"video/webm":       true,
	"video/ogg":        true,
	"application/json": true,
	"application/yaml": true,
}

func isSafeInlineMediaType(rawMediaType string) (bool, error) {
	rawMediaType, _, err := mime.ParseMediaType(rawMediaType)
	if err != nil {
		return false, fmt.Errorf("failed to parse media type %s: %w", rawMediaType, err)
	}

	return safeInlineMediaTypes[rawMediaType], nil
}

// detectMediaType finds out which mime type a file is.
// Will always return a mime type.
func detectMediaType(file *os.File) (string, error) {
	ext := filepath.Ext(file.Name())
	rawMediaType := mime.TypeByExtension(ext)

	if rawMediaType != "" {
		return rawMediaType, nil
	}

	buffer := make([]byte, 512)
	n, _ := io.ReadFull(file, buffer)

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "application/octet-stream", fmt.Errorf("failed to reset file reader for %s: %w", file.Name(), err)
	}

	return http.DetectContentType(buffer[:n]), nil
}
