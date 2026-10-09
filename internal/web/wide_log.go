package web

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"slices"
	"sync"
	"time"
)

type wideEventKey struct{}

// WideEvent holds attributes collected during request execution.
type WideEvent struct {
	mu    sync.Mutex
	attrs []slog.Attr
}

var importantEventKeys = map[string]bool{
	"error": true,
	"panic": true,
}

func (e *WideEvent) isImportantEvent() bool {
	return slices.ContainsFunc(e.attrs, func(attr slog.Attr) bool {
		return importantEventKeys[attr.Key]
	})
}

// AddEventAttrs appends one or more attributes to the request's wide log.
func AddEventAttrs(ctx context.Context, attrs ...slog.Attr) {
	if wl, ok := ctx.Value(wideEventKey{}).(*WideEvent); ok {
		wl.mu.Lock()
		defer wl.mu.Unlock()
		wl.attrs = append(wl.attrs, attrs...)
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("failed to write response writer for wide log: %w", err)
	}
	rw.bytesWritten += int64(n)
	return n, nil
}

// Unwrap allows access to the underlying ResponseWriter, which could be needed by http.ResponseController.
func (rw *responseWriterWrapper) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// ReadFrom needs to be implemented to satisfy the io.ReaderFrom interface, which is used by http.ServeContent for optimized file serving.
func (rw *responseWriterWrapper) ReadFrom(r io.Reader) (int64, error) {
	// If no status code has been set yet, use 200 OK.
	// This mirrors the behavior of the standard library.
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	// Use io.ReaderFrom for the optimized kernel transfer (zero-copy fast path).
	if rf, ok := rw.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(r)
		rw.bytesWritten += n
		return n, fmt.Errorf("failed to read from response writer for wide log: %w", err)
	}

	// Fallback to normal io.Copy when the fast path is not supported.
	n, err := io.Copy(rw.ResponseWriter, r)
	rw.bytesWritten += n
	return n, fmt.Errorf("failed to io copy for wide log: %w", err)
}

// WideEventMiddleware wraps around all Handler and ensures logging is done in the end
func WideEventMiddleware(next http.Handler, sampleRate float64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		we := &WideEvent{}

		ctx := context.WithValue(r.Context(), wideEventKey{}, we)
		r = r.WithContext(ctx)

		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			// Handle panics while ensuring the wide log is still emitted
			if rec := recover(); rec != nil {
				wrapped.statusCode = http.StatusInternalServerError
				AddEventAttrs(ctx, slog.Any("panic", rec))
				http.Error(wrapped, "Internal Server Error", http.StatusInternalServerError)
			}

			AddEventAttrs(ctx, slog.String("http.method", r.Method),
				slog.String("http.path", r.URL.Path),
				slog.Int("http.status_code", wrapped.statusCode),
				slog.Int64("http.bytes_written", wrapped.bytesWritten),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.String("user_agent", r.UserAgent()))

			var level slog.Level
			switch {
			case we.isImportantEvent():
				level = slog.LevelError
			case wrapped.statusCode >= 500:
				level = slog.LevelError
			case wrapped.statusCode >= 400:
				level = slog.LevelWarn
			default:
				level = slog.LevelInfo
			}

			if shouldSample(we, wrapped.statusCode, sampleRate) {
				slog.LogAttrs(ctx, level, "request_completed", we.attrs...)
			}
		}()
		next.ServeHTTP(wrapped, r)
	})
}

func shouldSample(we *WideEvent, statusCode int, sampleRate float64) bool {
	if we.isImportantEvent() {
		return true
	}

	if statusCode >= 400 {
		return true
	}
	//nolint:gosec // for log sampling this must not be super secure
	return rand.Float64() < sampleRate
}
