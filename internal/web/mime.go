package web

import (
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

func isSafeInlineMediaType(rawMediaType string) bool {
	parsedMediaType, _, err := mime.ParseMediaType(rawMediaType)
	if err != nil {
		return false
	}
	return safeInlineMediaTypes[parsedMediaType]
}

// detectMediaType finds out which mime type a file is.
// Will always return a mime type.
func detectMediaType(file *os.File) string {
	ext := filepath.Ext(file.Name())
	rawMediaType := mime.TypeByExtension(ext)

	if rawMediaType != "" {
		return rawMediaType
	}

	buffer := make([]byte, 512)
	n, _ := io.ReadFull(file, buffer)

	if n == 0 {
		return "application/octet-stream"
	}

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "application/octet-stream"
	}

	return http.DetectContentType(buffer[:n])
}
