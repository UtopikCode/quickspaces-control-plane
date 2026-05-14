package api

import (
	"net/http"
	"strings"
)

type Router struct {
	handler *Handler
}

// NewRouter creates a pure API router focused on core functionality.
// API documentation is served separately via Scalar.
func NewRouter(handler *Handler) http.Handler {
	router := &Router{handler: handler}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", router.handleHealth)
	mux.HandleFunc("/api/v1/auth/login", router.handleLogin)
	mux.HandleFunc("/api/v1/auth/callback", router.handleAuthCallback)
	mux.HandleFunc("/api/v1/auth/token", router.handleTokenExchange)
	mux.HandleFunc("/api/v1/workspaces", router.handleWorkspaces)
	mux.HandleFunc("/api/v1/workspaces/", router.handleWorkspaceActions)
	mux.HandleFunc("/api/v1/access", router.handleAccess)
	mux.HandleFunc("/api/v1/access/", router.handleAccess)
	return mux
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.NotFound(w, req)
		return
	}
	r.handler.Health(w, req)
}

func (r *Router) handleWorkspaces(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		r.handler.CreateWorkspace(w, req)
	case http.MethodGet:
		r.handler.ListWorkspaces(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleWorkspaceActions(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/v1/workspaces/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		http.NotFound(w, req)
		return
	}

	id := parts[0]
	if len(parts) == 1 {
		if req.Method == http.MethodGet {
			r.handler.GetWorkspace(w, req, id)
			return
		}
		http.NotFound(w, req)
		return
	}

	if req.Method != http.MethodPost {
		http.NotFound(w, req)
		return
	}

	switch parts[1] {
	case "start":
		r.handler.StartWorkspace(w, req, id)
	case "stop":
		r.handler.StopWorkspace(w, req, id)
	case "reconcile":
		r.handler.ReconcileWorkspace(w, req, id)
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleAccess(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/v1/access")
	path = strings.Trim(path, "/")

	if path != "" {
		http.NotFound(w, req)
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.handler.ListAccess(w, req)
	case http.MethodPost:
		r.handler.GrantAccess(w, req)
	case http.MethodDelete:
		r.handler.RemoveAccess(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.NotFound(w, req)
		return
	}
	r.handler.Login(w, req)
}

func (r *Router) handleAuthCallback(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handler.AuthCallback(w, req)
	case http.MethodPost:
		r.handler.TokenExchange(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *Router) handleTokenExchange(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.NotFound(w, req)
		return
	}
	r.handler.TokenExchange(w, req)
}
