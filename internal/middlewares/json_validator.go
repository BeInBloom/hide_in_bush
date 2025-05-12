package middlewares

import (
	"bytes"
	"io"
	"net/http"

	errorreporter "github.com/BeInBloom/hide_in_bush/internal/error_reporter"
	"github.com/BeInBloom/hide_in_bush/internal/validator"
)

func (m *Mw) BodyValidator(status int, v validator.Validator[[]byte]) chiMiddleware {
	reporter := errorreporter.New(nil)
	return func(next http.Handler) http.Handler {
		logger := m.logger.With("middleware", "validator")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				logger.Error("Failed to read request body", "error", err)
				m.handleJSONError(
					w, status, "Failed to read request body")
				return
			}

			if ok, err := v.Validate(data); err != nil {
				m.handleJSONError(w, status, reporter.Report(err)...)
				return
			} else if !ok {
				logger.Error("Request body is invalid")
				m.handleJSONError(w, status, "Request body is invalid")
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(data))
			next.ServeHTTP(w, r)
		})
	}
}
