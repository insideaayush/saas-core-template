package errorreporting

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
)

type Middleware struct {
	reporter Reporter
}

func NewMiddleware(reporter Reporter) Middleware {
	return Middleware{reporter: reporter}
}

func (m Middleware) Wrap(next http.Handler) http.Handler {
	if m.reporter == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			if recovered := recover(); recovered != nil {
				err := fmt.Errorf("panic: %v", recovered)
				m.reporter.CaptureException(r.Context(), err, map[string]string{
					"http.method": r.Method,
					"http.path":   r.URL.Path,
					"panic":       "true",
				})

				// Don't leak panic details to clients.
				http.Error(sw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if sw.status >= 500 {
				m.reporter.CaptureException(r.Context(), fmt.Errorf("server error %d", sw.status), map[string]string{
					"http.method": r.Method,
					"http.path":   r.URL.Path,
					"http.status": fmt.Sprintf("%d", sw.status),
				})
			}
		}()

		next.ServeHTTP(sw, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	// If a handler never explicitly calls WriteHeader, net/http will write 200 on first write.
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	stack := strings.TrimSpace(string(debug.Stack()))
	if stack == "" {
		return err
	}
	return fmt.Errorf("%w\n%s", err, stack)
}

func Capture(ctx context.Context, reporter Reporter, err error, attrs map[string]string) {
	if reporter == nil || err == nil {
		return
	}
	reporter.CaptureException(ctx, err, attrs)
}
