package api

import (
	"net/http"
	"strings"

	_ "github.com/UtopikCode/quickspaces-control-plane/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) http.Handler {
	router := &Router{handler: handler}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", router.handleHealth)
	mux.HandleFunc("/api/v1/workspaces", router.handleWorkspaces)
	mux.HandleFunc("/api/v1/workspaces/", router.handleWorkspaceActions)
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently).ServeHTTP(w, r)
	})
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently).ServeHTTP(w, r)
	})
	mux.HandleFunc("/docs/", func(w http.ResponseWriter, r *http.Request) {
		http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently).ServeHTTP(w, r)
	})
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
