package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/UtopikCode/quickspaces-control-plane/application"
	"github.com/UtopikCode/quickspaces-control-plane/domain"
	"github.com/UtopikCode/quickspaces-control-plane/internal/application/auth"
	githubclient "github.com/UtopikCode/quickspaces-control-plane/internal/infrastructure/github"
)

// errorResponse defines the shape of API error responses.
type errorResponse struct {
	Error string `json:"error"`
}

type GrantAccessRequestPayload struct {
	SubjectType string `json:"subjectType"`
	SubjectID   string `json:"subjectId"`
	Role        string `json:"role"`
}

type RemoveAccessRequestPayload struct {
	SubjectType string `json:"subjectType"`
	SubjectID   string `json:"subjectId"`
}

type AccessRuleResponse struct {
	SubjectType string `json:"subjectType"`
	SubjectID   string `json:"subjectId"`
	Role        string `json:"role"`
}

type Handler struct {
	service      *application.WorkspaceService
	authService  *auth.Service
	githubClient githubclient.GitHubClient
}

func NewHandler(service *application.WorkspaceService, authService *auth.Service, githubClient githubclient.GitHubClient) *Handler {
	return &Handler{service: service, authService: authService, githubClient: githubClient}
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
// @Param request body CreateWorkspaceRequestPayload true "Create workspace request"
// @Success 201 {object} domain.Workspace
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /workspaces [post]
func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	if err := h.authorizeRequest(r, "user"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

	var req CreateWorkspaceRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if req.Repo == "" || req.Owner == "" || req.Ref == "" {
		writeError(w, http.StatusBadRequest, "repo, owner, and ref are required")
		return
	}

	executionProfileBytes, err := json.Marshal(req.ExecutionProfile)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid executionProfile payload")
		return
	}

	workspace, err := h.service.CreateWorkspace(r.Context(), application.CreateWorkspaceRequest{
		Repo:             req.Repo,
		Owner:            req.Owner,
		Ref:              req.Ref,
		ExecutionProfile: domain.ExecutionProfile(executionProfileBytes),
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
	if err := h.authorizeRequest(r, "user"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

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
	if err := h.authorizeRequest(r, "user"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

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
	if err := h.authorizeRequest(r, "user"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

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
	if err := h.authorizeRequest(r, "user"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

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
	if err := h.authorizeRequest(r, "user"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

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

// GrantAccess godoc
// @Summary Grant or update an access rule
// @Description Inserts or updates a global GitHub access rule.
// @Tags Access
// @Accept json
// @Produce json
// @Param request body GrantAccessRequestPayload true "Grant access request"
// @Success 200 {object} AccessRuleResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /access [post]
func (h *Handler) GrantAccess(w http.ResponseWriter, r *http.Request) {
	if err := h.authorizeRequest(r, "admin"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

	var req GrantAccessRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if req.SubjectType == "" || req.SubjectID == "" || req.Role == "" {
		writeError(w, http.StatusBadRequest, "subjectType, subjectId, and role are required")
		return
	}

	if err := h.authService.GrantAccess(r.Context(), req.SubjectType, req.SubjectID, req.Role); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AccessRuleResponse(req))
}

// ListAccess godoc
// @Summary List access rules
// @Description Returns all configured GitHub access rules.
// @Tags Access
// @Produce json
// @Success 200 {array} AccessRuleResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /access [get]
func (h *Handler) ListAccess(w http.ResponseWriter, r *http.Request) {
	if err := h.authorizeRequest(r, "admin"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

	rules, err := h.authService.ListAccess(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]AccessRuleResponse, 0, len(rules))
	for _, rule := range rules {
		response = append(response, AccessRuleResponse{SubjectType: rule.SubjectType, SubjectID: rule.SubjectID, Role: rule.Role})
	}
	writeJSON(w, http.StatusOK, response)
}

// RemoveAccess godoc
// @Summary Remove an access rule
// @Description Removes a global GitHub access rule.
// @Tags Access
// @Accept json
// @Produce json
// @Param request body RemoveAccessRequestPayload true "Remove access request"
// @Success 204
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /access [delete]
func (h *Handler) RemoveAccess(w http.ResponseWriter, r *http.Request) {
	if err := h.authorizeRequest(r, "admin"); err != nil {
		writeError(w, getStatusCode(err), err.Error())
		return
	}

	var req RemoveAccessRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if req.SubjectType == "" || req.SubjectID == "" {
		writeError(w, http.StatusBadRequest, "subjectType and subjectId are required")
		return
	}

	if err := h.authService.RemoveAccess(r.Context(), req.SubjectType, req.SubjectID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) authorizeRequest(r *http.Request, requiredRole string) error {
	user, orgs, teams, err := h.authenticateGitHubRequest(r)
	if err != nil {
		return err
	}

	allowed, role, err := h.authService.Authorize(r.Context(), user, orgs, teams)
	if err != nil {
		return err
	}
	if !allowed {
		return &apiError{status: http.StatusForbidden, message: "access denied"}
	}

	if !hasRequiredRole(role, requiredRole) {
		return &apiError{status: http.StatusForbidden, message: "insufficient role"}
	}
	return nil
}

func (h *Handler) authenticateGitHubRequest(r *http.Request) (githubclient.GithubUser, []string, []githubclient.GithubTeam, error) {
	token, ok := parseBearerToken(r.Header.Get("Authorization"))
	if !ok {
		return githubclient.GithubUser{}, nil, nil, &apiError{status: http.StatusUnauthorized, message: "missing or invalid authorization header"}
	}

	user, err := h.githubClient.GetUser(token)
	if err != nil {
		return githubclient.GithubUser{}, nil, nil, &apiError{status: http.StatusUnauthorized, message: "invalid access token"}
	}

	orgs, err := h.githubClient.GetUserOrgs(token)
	if err != nil {
		return githubclient.GithubUser{}, nil, nil, &apiError{status: http.StatusUnauthorized, message: "invalid access token"}
	}

	teams, err := h.githubClient.GetUserTeams(token)
	if err != nil {
		return githubclient.GithubUser{}, nil, nil, &apiError{status: http.StatusUnauthorized, message: "invalid access token"}
	}

	return user, orgs, teams, nil
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

type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string {
	return e.message
}

func getStatusCode(err error) int {
	var apiErr *apiError
	if errors.As(err, &apiErr) {
		return apiErr.status
	}
	return http.StatusInternalServerError
}

// Login godoc
// @Summary Redirect user to GitHub OAuth login
// @Description Redirects a client to GitHub to begin OAuth authentication.
// @Tags Auth
// @Produce json
// @Success 302
// @Router /auth/login [get]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if h.githubClient == nil {
		writeError(w, http.StatusInternalServerError, "OAuth not configured")
		return
	}
	http.Redirect(w, r, h.githubClient.AuthorizeURL(), http.StatusFound)
}

// AuthCallback godoc
// OAuth2 callback handler - receives authorization code from GitHub.
// This endpoint is called by GitHub's OAuth redirect. It doesn't exchange the code
// here; instead, it sends the code back to Scalar which will exchange it via TokenExchange.
func (h *Handler) AuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		// GitHub OAuth denied or errored
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Authorization Denied</title></head>
<body>
<script>
if (window.opener) {
  window.opener.postMessage({
    type: 'oauth_error',
    error: 'User denied authorization'
  }, window.location.origin);
  window.close();
}
</script>
</body>
</html>`)); err != nil {
			log.Printf("failed to write auth denied response: %v", err)
		}
		return
	}

	if code == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Authorization Error</title></head>
<body>
<script>
if (window.opener) {
  window.opener.postMessage({
    type: 'oauth_error',
    error: 'No authorization code received'
  }, window.location.origin);
  window.close();
}
</script>
</body>
</html>`)); err != nil {
			log.Printf("failed to write auth error response: %v", err)
		}
		return
	}

	log.Printf("[AuthCallback] Received code from GitHub: %s (first 10 chars)", code[:min(10, len(code))])

	// Don't exchange the code here! Just send it back to Scalar.
	// Scalar will POST the code to /api/v1/auth/token to exchange it.
	// This way, the code is only used once.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authorization Successful</title></head>
<body>
<p>Completing authorization...</p>
<script>
if (window.opener) {
  // Send the authorization code back to Scalar popup handler
  // Scalar will use this code to POST to /api/v1/auth/token
  window.opener.postMessage({
    type: 'oauth_code',
    code: '%s',
    state: '%s'
  }, window.location.origin);
  window.close();
} else {
  document.body.innerHTML = '<h1>Authorization received</h1><p>You can close this window.</p>';
}
</script>
</body>
</html>`, code, state); err != nil {
		log.Printf("failed to write auth success response: %v", err)
	}
}

// TokenExchange godoc
// @Summary Exchange OAuth code for token
// @Description Exchanges an OAuth2 code for an access token (used by client-side apps).
// The code_verifier is optional and used for PKCE-protected flows.
// @Tags Auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param code formData string true "OAuth authorization code"
// @Param code_verifier formData string false "PKCE code verifier (if PKCE was used)"
// @Param grant_type formData string false "Grant type (authorization_code)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errorResponse
// @Router /auth/token [post]
func (h *Handler) TokenExchange(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")

	log.Printf("[TokenExchange] Received: code=%s (first 10 chars), has_code_verifier=%v",
		code[:min(10, len(code))], codeVerifier != "")

	if code == "" {
		writeError(w, http.StatusBadRequest, "code parameter required")
		return
	}

	if h.githubClient == nil {
		writeError(w, http.StatusInternalServerError, "OAuth not configured")
		return
	}

	token, err := h.githubClient.ExchangeCode(code, codeVerifier)
	if err != nil {
		log.Printf("[TokenExchange] Token exchange failed: %v", err)
		writeError(w, http.StatusBadRequest, "failed to exchange code: "+err.Error())
		return
	}

	log.Printf("[TokenExchange] Token exchange successful")
	writeJSON(w, http.StatusOK, map[string]string{
		"access_token": token,
		"token_type":   "bearer",
	})
}

// CreateWorkspaceRequestPayload models the payload for workspace creation.
// swagger:model
type CreateWorkspaceRequestPayload struct {
	Repo             string                 `json:"repo"`
	Owner            string                 `json:"owner"`
	Ref              string                 `json:"ref"`
	ExecutionProfile map[string]interface{} `json:"executionProfile"`
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
