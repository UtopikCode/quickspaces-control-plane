package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/UtopikCode/quickspaces-control-plane/application"
	"github.com/UtopikCode/quickspaces-control-plane/domain"
)

// errorResponse defines the shape of API error responses.
type errorResponse struct {
	Error string `json:"error"`
}

type Handler struct {
	service *application.WorkspaceService
}

func NewHandler(service *application.WorkspaceService) *Handler {
	return &Handler{service: service}
}

// Health godoc
// @Summary Health check
// @Description Returns whether the control plane API is healthy.
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateWorkspace godoc
// @Summary Create a new workspace
// @Description Creates a workspace with a desired state and execution profile.
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param request body createWorkspaceRequest true "Create workspace request"
// @Success 201 {object} domain.Workspace
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /workspaces [post]
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

// ListWorkspaces godoc
// @Summary List workspaces
// @Description Returns all configured workspaces.
// @Tags Workspaces
// @Produce json
// @Success 200 {array} domain.Workspace
// @Failure 500 {object} errorResponse
// @Router /workspaces [get]
func (h *Handler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	workspaces, err := h.service.ListWorkspaces(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, workspaces)
}

// GetWorkspace godoc
// @Summary Get a workspace
// @Description Returns the workspace identified by its ID.
// @Tags Workspaces
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} domain.Workspace
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /workspaces/{id} [get]
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

// StartWorkspace godoc
// @Summary Start a workspace
// @Description Transitions the workspace to the running state.
// @Tags Workspaces
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} domain.Workspace
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /workspaces/{id}/start [post]
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

// StopWorkspace godoc
// @Summary Stop a workspace
// @Description Transitions the workspace to the stopped state.
// @Tags Workspaces
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} domain.Workspace
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /workspaces/{id}/stop [post]
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

// ReconcileWorkspace godoc
// @Summary Reconcile a workspace
// @Description Reconciles the configured workspace state with the execution environment.
// @Tags Workspaces
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} domain.Workspace
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /workspaces/{id}/reconcile [post]
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

// createWorkspaceRequest models the payload for workspace creation.
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
	writeJSON(w, status, errorResponse{Error: message})
}
