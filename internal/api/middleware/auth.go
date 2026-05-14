package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	githubclient "github.com/UtopikCode/quickspaces-control-plane/internal/infrastructure/github"
)

type contextKey string

const (
	userContextKey contextKey = "auth.user"
	roleContextKey contextKey = "auth.role"
)

type AuthMiddleware struct {
	githubClient githubclient.GitHubClient
}

func NewAuthMiddleware(githubClient githubclient.GitHubClient) *AuthMiddleware {
	return &AuthMiddleware{githubClient: githubClient}
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := parseBearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}

		user, err := m.githubClient.GetUser(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid access token")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := RoleFromContext(r.Context())
			if !ok || !hasRequiredRole(role, requiredRole) {
				writeError(w, http.StatusForbidden, "insufficient role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleContextKey).(string)
	return role, ok
}

func HasRole(ctx context.Context, required string) bool {
	role, ok := RoleFromContext(ctx)
	return ok && hasRequiredRole(role, required)
}

func UserFromContext(ctx context.Context) (githubclient.GithubUser, bool) {
	user, ok := ctx.Value(userContextKey).(githubclient.GithubUser)
	return user, ok
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleContextKey, role)
}

func parseBearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

func hasRequiredRole(current, required string) bool {
	priority := map[string]int{"user": 1, "admin": 2}
	return priority[strings.ToLower(current)] >= priority[strings.ToLower(required)]
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
