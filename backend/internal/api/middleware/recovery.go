package middleware

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func Recovery(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.WithFields(logrus.Fields{
						"panic":      rec,
						"request_id": GetRequestID(r.Context()),
					}).Error("panic recovered")
					http.Error(w, `{"error":{"code":"INTERNAL","message":"internal server error"}}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
