package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
)

const (
	maxBodySize = 1 << 17 // 128 KB
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
	body   *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if len(b) <= maxBodySize {
		r.body.Write(b)
	}
	r.size += size
	return size, err
}

func (m *Mw) Logger() chiMiddleware {
	logger := m.logger.With("middleware", "logger")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			requestID := middleware.GetReqID(r.Context())

			clientIP := getClientIP(r)

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("Failed to read request body", "error", err)
				m.handleJSONError(w, http.StatusBadRequest, "Failed to read request body")
				return
			}
			r.Body.Close()

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			//Стоит понимать, что тут может логироваться и личная информация
			//Что не жалетельно, потом по запросу можно устать ее удалять
			logger.Info("Request received",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_ip", clientIP,
				"user_agent", r.UserAgent(),
				"body", string(body),
			)

			recorder := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(recorder, r)

			responseBody := recorder.body.String()

			respLogAttrs := []any{
				"request_id", requestID,
				"method", r.Method,
				"status", recorder.status,
				"size", recorder.size,
				"body", responseBody,
				"duration", time.Since(startTime).Milliseconds(),
			}

			switch {
			case recorder.status >= 500:
				logger.Error("Response error", respLogAttrs...)
			case recorder.status >= 400:
				logger.Warn("Response warning", respLogAttrs...)
			default:
				logger.Info("Response sent", respLogAttrs...)
			}
		})
	}
}

func getClientIP(r *http.Request) string {
	clientIP := r.RemoteAddr
	if realIP := r.Header.Get("X-Forwarded-For"); realIP != "" {
		clientIP = realIP
	} else if forwardedFor := r.Header.Get("X-Real-IP"); forwardedFor != "" {
		//если верить нейронки, то он вернет цепочку ip резделенную ", "
		//если исходникам, то он вернет только первый ip
		//поверю исходникам chi мидлтвари
		clientIP = forwardedFor
	}

	return clientIP
}
