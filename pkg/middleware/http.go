package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xyun1996/social_backend/pkg/auth"
	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/pkg/metrics"
	"github.com/xyun1996/social_backend/pkg/transport"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	n, err := r.ResponseWriter.Write(body)
	r.bytes += n
	return n, err
}

func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func TraceIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(traceIDKey).(string)
	return value
}

func WithRequestContext(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
			if requestID == "" {
				requestID = mustToken()
			}

			traceID := strings.TrimSpace(r.Header.Get("X-Trace-ID"))
			if traceID == "" {
				traceID = requestID
			}

			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("X-Trace-ID", traceID)

			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			ctx = context.WithValue(ctx, traceIDKey, traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("request panic recovered",
						slog.Any("panic", recovered),
						slog.String("request_id", RequestIDFromContext(r.Context())),
						slog.String("trace_id", TraceIDFromContext(r.Context())),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("stack", string(debug.Stack())),
					)
					transport.WriteError(w, apperrors.Internal())
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func AccessLog(logger *slog.Logger, registry *metrics.Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &responseRecorder{ResponseWriter: w}
			if registry != nil {
				registry.IncInflight()
				defer registry.DecInflight()
			}

			next.ServeHTTP(recorder, r)

			route := r.Pattern
			if route == "" {
				route = r.URL.Path
			}
			if recorder.status == 0 {
				recorder.status = http.StatusOK
			}

			duration := time.Since(started)
			if registry != nil {
				registry.Record(r.Method, route, recorder.status, recorder.bytes, duration)
			}

			logger.Info("http request completed",
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("trace_id", TraceIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("route", route),
				slog.Int("status", recorder.status),
				slog.Int("bytes", recorder.bytes),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.String("remote_addr", clientIP(r)),
			)
		})
	}
}

func RequireInternalToken(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if strings.TrimSpace(token) == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/v1/internal/") {
				next.ServeHTTP(w, r)
				return
			}

			if auth.InternalTokenFromRequest(r) != token {
				err := apperrors.New("unauthorized", "internal service token is required", http.StatusUnauthorized)
				transport.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireOpsToken(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if strings.TrimSpace(token) == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/v1/ops/") {
				next.ServeHTTP(w, r)
				return
			}

			if auth.BearerToken(r) != token {
				err := apperrors.New("unauthorized", "ops bearer token is required", http.StatusUnauthorized)
				transport.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AuditLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &responseRecorder{ResponseWriter: w}
			next.ServeHTTP(recorder, r)

			if !shouldAudit(r, recorder.status) {
				return
			}

			logger.Info("audit event",
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("trace_id", TraceIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", recorder.status),
				slog.String("actor_player_id", r.Header.Get("X-Actor-Player-ID")),
				slog.String("operator_id", r.Header.Get("X-Operator-ID")),
				slog.String("remote_addr", clientIP(r)),
			)
		})
	}
}

type rateLimiter struct {
	rps    int
	burst  int
	mu     sync.Mutex
	window map[string]bucket
}

type bucket struct {
	second int64
	count  int
}

func RateLimit(rps int, burst int) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		rps:    rps,
		burst:  burst,
		window: make(map[string]bucket),
	}

	return func(next http.Handler) http.Handler {
		if rps <= 0 {
			return next
		}

		if burst <= 0 {
			limiter.burst = rps
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exemptFromRateLimit(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := clientIP(r) + "|" + r.Method
			if !limiter.allow(key, time.Now().Unix()) {
				err := apperrors.New("rate_limited", "request rate exceeded", http.StatusTooManyRequests)
				w.Header().Set("Retry-After", "1")
				transport.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (l *rateLimiter) allow(key string, second int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.window[key]
	if entry.second != second {
		entry = bucket{second: second}
	}

	limit := l.rps + l.burst
	if entry.count >= limit {
		l.window[key] = entry
		return false
	}

	entry.count++
	l.window[key] = entry
	return true
}

func mustToken() string {
	value, err := idgen.Token(8)
	if err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return value
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}

func exemptFromRateLimit(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}

	return strings.HasPrefix(r.URL.Path, "/v1/internal/") || r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/metrics"
}

func shouldAudit(r *http.Request, status int) bool {
	if status < 200 || status >= 300 {
		return false
	}

	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return false
	}

	path := r.URL.Path
	return strings.HasPrefix(path, "/v1/internal/") ||
		strings.HasPrefix(path, "/v1/ops/") ||
		strings.Contains(path, "/governance") ||
		strings.Contains(path, "/moderators") ||
		strings.Contains(path, "/mutes") ||
		strings.Contains(path, "/announcement") ||
		strings.Contains(path, "/kick") ||
		strings.Contains(path, "/transfer-owner") ||
		strings.Contains(path, "/activities/")
}
