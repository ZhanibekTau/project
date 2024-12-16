package api

import (
	"log"
	"net/http"
	"project/internal/consts"
	"time"
)

// Логирующий миддлвар
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		// Determine color for the method
		methodColor := consts.Green // Default for successful status
		if lrw.statusCode >= 400 && lrw.statusCode < 500 {
			methodColor = consts.Yellow // Client error
		} else if lrw.statusCode >= 500 {
			methodColor = consts.Red // Server error
		}

		// Format the log
		log.Printf("%s[%s%s%s%s] \"%s\" - status - %s%d%s, size %d bytes in %v second%s",
			consts.Cyan,                         // Timestamp prefix color
			methodColor, r.Method, consts.Reset, // Colored method
			consts.Cyan, r.URL.Path, // URL path in quotes
			methodColor, lrw.statusCode, consts.Reset, // Colored status
			lrw.size, time.Since(start).Seconds(), consts.Reset) // Execution time
	})
}

// Структура для логирования ответа
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// Переопределение метода WriteHeader для захвата статуса
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Переопределение метода Write для захвата размера контента
func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}
