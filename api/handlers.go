package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/UtopikCode/quickspaces-control-plane/application"
	"github.com/UtopikCode/quickspaces-control-plane/domain"
)

type Handler struct {
	service *application.WorkspaceService
}

func NewHandler(service *application.WorkspaceService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if req.Repo == "" || req.Owner == "" || req.Ref == "" {
		writeError(w, http.StatusBadRequest, "repo, owner, and ref are required")
		return
	}

	workspace, err := h.service.CreateWorkspace(r.Context(), application.CreateWorkspaceRequest{
		Repo:             req.Repo,
		Owner:            req.Owner,
		Ref:              req.Ref,
		ExecutionProfile: req.ExecutionProfile,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, workspace)
}

func (h *Handler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	workspaces, err := h.service.ListWorkspaces(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, workspaces)
}

func (h *Handler) GetWorkspace(w http.ResponseWriter, r *http.Request, id string) {
	workspace, err := h.service.GetWorkspace(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			writeError(w, http.StatusNotFound, "workspace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) StartWorkspace(w http.ResponseWriter, r *http.Request, id string) {
	workspace, err := h.service.StartWorkspace(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			writeError(w, http.StatusNotFound, "workspace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) StopWorkspace(w http.ResponseWriter, r *http.Request, id string) {
	workspace, err := h.service.StopWorkspace(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			writeError(w, http.StatusNotFound, "workspace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) ReconcileWorkspace(w http.ResponseWriter, r *http.Request, id string) {
	workspace, err := h.service.ReconcileWorkspace(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			writeError(w, http.StatusNotFound, "workspace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

type createWorkspaceRequest struct {
	Repo             string                  `json:"repo"`
	Owner            string                  `json:"owner"`
	Ref              string                  `json:"ref"`
	ExecutionProfile domain.ExecutionProfile `json:"executionProfile"`
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
