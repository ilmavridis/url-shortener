package middleware

import (
	"ilmavridis/url-shortener/logger"

	"net/http"
	"time"

	"go.uber.org/zap"
)

// It logs the status code of each request by creating a new type that implements ResponseWriter
// It also logs the response time of each request
func Logger(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()

		recorder := &ResponseRecorder{
			ResponseWriter: w,
			Status:         200,
		}

		h.ServeHTTP(recorder, r) // Inject our implementation of http.ResponseWriter

		uri := zap.String("url", r.URL.Path)
		method := zap.String("method", r.Method)
		status := zap.Int("status", recorder.Status)
		duration := zap.Duration("duration", time.Duration(time.Since(startTime)))

		if recorder.Status >= 400 && recorder.Status < 500 {
			logger.HttpWarn("Client error", method, uri, status)
		} else if recorder.Status >= 500 {
			logger.HttpError("Internal error", method, uri, status)
		} else {
			logger.Info("Request received", method, uri, duration, status)
		}

	})
}

type ResponseRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *ResponseRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}
