package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

func Logging(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			log.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      sw.status,
				"duration_ms": time.Since(start).Milliseconds(),
				"request_id":  GetRequestID(r.Context()),
			}).Info("request")
		})
	}
}
