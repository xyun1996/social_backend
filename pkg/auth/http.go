package auth

import (
	"net/http"
	"os"
	"strings"
)

// BearerToken extracts a bearer token from the Authorization header.
func BearerToken(r *http.Request) string {
	if r == nil {
		return ""
	}

	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return ""
	}

	return strings.TrimSpace(header[7:])
}

// InternalTokenFromRequest extracts an internal service token from the request.
func InternalTokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	if token := strings.TrimSpace(r.Header.Get("X-Internal-Token")); token != "" {
		return token
	}

	return BearerToken(r)
}

// ApplyInternalToken writes the configured internal service token to the request if present.
func ApplyInternalToken(req *http.Request) {
	if req == nil {
		return
	}

	token := strings.TrimSpace(os.Getenv("APP_INTERNAL_TOKEN"))
	if token == "" {
		return
	}

	req.Header.Set("X-Internal-Token", token)
}
