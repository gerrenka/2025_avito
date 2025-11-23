package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"review-service/internal/handlers"
	"review-service/internal/handlers/middleware"
	"review-service/internal/repository/postgres"
	"review-service/internal/service"
)

type TestServer struct {
	Server *httptest.Server
	DB     *sql.DB
}

func setupTestServer(t *testing.T) *TestServer {

	db, err := sql.Open("postgres", "host=localhost port=5433 user=reviewuser password=reviewpass dbname=reviewdb_test sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	cleanupDB(t, db)
	setupSchema(t, db)

	userRepo := postgres.NewUserPostgresRepository(db)
	teamRepo := postgres.NewTeamPostgresRepository(db)
	prRepo := postgres.NewPullRequestPostgresRepository(db)

	userService := service.NewUserService(userRepo, prRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	router := handlers.NewRouter(userService, teamService, prService)
	mux := http.NewServeMux()
	router.SetupRoutes(mux)

	server := httptest.NewServer(middleware.Logger(middleware.Recovery(mux)))

	return &TestServer{
		Server: server,
		DB:     db,
	}
}

func cleanupDB(t *testing.T, db *sql.DB) {
	queries := []string{
		"DROP TABLE IF EXISTS pr_reviewers",
		"DROP TABLE IF EXISTS pull_requests",
		"DROP TABLE IF EXISTS users",
		"DROP TABLE IF EXISTS teams",
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			t.Logf("Warning during cleanup: %v", err)
		}
	}
}

func setupSchema(t *testing.T, db *sql.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS teams (
		team_name VARCHAR(255) PRIMARY KEY,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS users (
		user_id VARCHAR(255) PRIMARY KEY,
		username VARCHAR(255) NOT NULL,
		team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);

	CREATE TABLE IF NOT EXISTS pull_requests (
		pull_request_id VARCHAR(255) PRIMARY KEY,
		pull_request_name VARCHAR(500) NOT NULL,
		author_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
		status VARCHAR(50) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		merged_at TIMESTAMP NULL
	);

	CREATE TABLE IF NOT EXISTS pr_reviewers (
		pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
		user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
		PRIMARY KEY (pull_request_id, user_id)
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to setup schema: %v", err)
	}
}

func (ts *TestServer) Close() {
	ts.Server.Close()
	ts.DB.Close()
}

func doRequest(t *testing.T, method, url string, body interface{}) (*http.Response, map[string]interface{}) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to do request: %v", err)
	}
	defer resp.Body.Close()

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return resp, responseData
}

func TestCreateAndGetTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := setupTestServer(t)
	defer ts.Close()

	baseURL := ts.Server.URL

	teamData := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}

	resp, data := doRequest(t, "POST", baseURL+"/team/add", teamData)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	team := data["team"].(map[string]interface{})
	if team["team_name"] != "backend" {
		t.Errorf("Expected team_name 'backend', got %v", team["team_name"])
	}

	resp, data = doRequest(t, "GET", baseURL+"/team/get?team_name=backend", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if data["team_name"] != "backend" {
		t.Errorf("Expected team_name 'backend', got %v", data["team_name"])
	}

	members := data["members"].([]interface{})
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}
}

func TestCreatePullRequestWithAutoAssignment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := setupTestServer(t)
	defer ts.Close()

	baseURL := ts.Server.URL

	teamData := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "Dave", "is_active": true},
		},
	}

	doRequest(t, "POST", baseURL+"/team/add", teamData)

	prData := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search feature",
		"author_id":         "u1",
	}

	resp, data := doRequest(t, "POST", baseURL+"/pullRequest/create", prData)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	pr := data["pr"].(map[string]interface{})

	if pr["status"] != "OPEN" {
		t.Errorf("Expected status 'OPEN', got %v", pr["status"])
	}

	reviewers := pr["assigned_reviewers"].([]interface{})
	if len(reviewers) == 0 {
		t.Error("Expected at least 1 reviewer to be assigned")
	}

	for _, reviewer := range reviewers {
		if reviewer == "u1" {
			t.Error("Author should not be assigned as reviewer")
		}
	}

	t.Logf("Auto-assigned %d reviewers: %v", len(reviewers), reviewers)
}

func TestMergePullRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := setupTestServer(t)
	defer ts.Close()

	baseURL := ts.Server.URL

	teamData := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	doRequest(t, "POST", baseURL+"/team/add", teamData)

	prData := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	doRequest(t, "POST", baseURL+"/pullRequest/create", prData)

	mergeData := map[string]interface{}{
		"pull_request_id": "pr-1001",
	}

	resp, data := doRequest(t, "POST", baseURL+"/pullRequest/merge", mergeData)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	pr := data["pr"].(map[string]interface{})

	if pr["status"] != "MERGED" {
		t.Errorf("Expected status 'MERGED', got %v", pr["status"])
	}

	if pr["mergedAt"] == nil {
		t.Error("Expected 'mergedAt' to be set after merge")
	}

	resp, data = doRequest(t, "POST", baseURL+"/pullRequest/merge", mergeData)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 (idempotent), got %d", resp.StatusCode)
	}

	pr = data["pr"].(map[string]interface{})
	if pr["status"] != "MERGED" {
		t.Errorf("Expected status still 'MERGED', got %v", pr["status"])
	}
}

func TestGetUserReviews(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := setupTestServer(t)
	defer ts.Close()

	baseURL := ts.Server.URL

	teamData := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	doRequest(t, "POST", baseURL+"/team/add", teamData)

	for i := 1; i <= 3; i++ {
		prData := map[string]interface{}{
			"pull_request_id":   "pr-100" + string(rune(i)),
			"pull_request_name": "Feature " + string(rune(i)),
			"author_id":         "u1",
		}
		doRequest(t, "POST", baseURL+"/pullRequest/create", prData)
	}

	resp, data := doRequest(t, "GET", baseURL+"/users/getReview?user_id=u2", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if data["user_id"] != "u2" {
		t.Errorf("Expected user_id 'u2', got %v", data["user_id"])
	}

	prs := data["pull_requests"].([]interface{})
	t.Logf("User u2 is assigned to %d PRs", len(prs))

	if len(prs) > 0 {
		firstPR := prs[0].(map[string]interface{})
		if firstPR["pull_request_id"] == nil {
			t.Error("Expected pull_request_id in PR")
		}
		if firstPR["status"] == nil {
			t.Error("Expected status in PR")
		}
	}
}
