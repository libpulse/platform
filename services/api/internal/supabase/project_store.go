package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

// Project structure for database operations
type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	OwnerUserID   string    `json:"owner_user_id"`
	RetentionDays *int      `json:"retention_days,omitempty"`
	SignedOnly    *bool     `json:"signed_only,omitempty"`
	IsDemo        *bool     `json:"is_demo,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProjectStore is a thin wrapper around Client that provides project-related data access.
// It is used as the concrete implementation injected into handlers.
type ProjectStore struct {
	Client *Client
}

// CreateProject => POST /rest/v1/projects
func (s *ProjectStore) CreateProject(ctx context.Context, name string, ownerUserID string) (*Project, error) {
	if name == "" {
		return nil, errors.New("project name cannot be empty")
	}
	if ownerUserID == "" {
		return nil, errors.New("owner user id cannot be empty")
	}

	url := s.Client.BaseRestURL + "/projects"

	payload := map[string]interface{}{
		"name":          name,
		"owner_user_id": ownerUserID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Supabase REST API headers
	req.Header.Set("apikey", s.Client.ServiceRoleKey)
	req.Header.Set("Authorization", "Bearer "+s.Client.ServiceRoleKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.Client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		log.Printf("supabase rest api error: status=%d body=%s", resp.StatusCode, bodyStr)

		// Return the full body in the error so handler can detect constraint violations
		return nil, errors.New(bodyStr)
	}

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, errors.New("no project returned from database")
	}

	return &projects[0], nil
}
