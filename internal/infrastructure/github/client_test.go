package githubclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientFetchesGitHubIdentityPerRequest(t *testing.T) {
	userRequests := 0
	orgRequests := 0
	teamRequests := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		userRequests++
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("expected Authorization header Bearer token, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(GithubUser{Login: "alice", ID: 1})
	})
	mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		orgRequests++
		_ = json.NewEncoder(w).Encode([]map[string]string{{"login": "acme"}})
	})
	mux.HandleFunc("/user/teams", func(w http.ResponseWriter, r *http.Request) {
		teamRequests++
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"name":         "developers",
				"organization": map[string]string{"login": "acme"},
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	oldBaseURL := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = oldBaseURL }()

	client := NewClient("id", "secret", "http://callback")
	client.httpClient = server.Client()

	for i := 0; i < 2; i++ {
		_, err := client.GetUser("token")
		if err != nil {
			t.Fatalf("GetUser failed: %v", err)
		}
		_, err = client.GetUserOrgs("token")
		if err != nil {
			t.Fatalf("GetUserOrgs failed: %v", err)
		}
		_, err = client.GetUserTeams("token")
		if err != nil {
			t.Fatalf("GetUserTeams failed: %v", err)
		}
	}

	if userRequests != 2 {
		t.Fatalf("expected 2 /user requests, got %d", userRequests)
	}
	if orgRequests != 2 {
		t.Fatalf("expected 2 /user/orgs requests, got %d", orgRequests)
	}
	if teamRequests != 2 {
		t.Fatalf("expected 2 /user/teams requests, got %d", teamRequests)
	}
}
