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

// ProjectKey structure for database operations
type ProjectKey struct {
	ID                string     `json:"id"`
	ProjectID         string     `json:"project_id"`
	Label             string     `json:"label"`
	Env               string     `json:"env"`
	SignedOnly        bool       `json:"signed_only"`
	PublicKey         string     `json:"public_key"`
	SecretEnc         string     `json:"secret_enc"`         // Stores AES-GCM encrypted secret (can be decrypted)
	SecretFingerprint string     `json:"secret_fingerprint"` // Stores HMAC-SHA256 hashed secret (one-way)
	Disabled          bool       `json:"disabled"`
	CreatedBy         string     `json:"created_by"`
	CreatedAt         time.Time  `json:"created_at"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
}

// GetProjectID implements ProjectKeyInfo interface
func (pk *ProjectKey) GetProjectID() string {
	return pk.ProjectID
}

// IsDisabled implements ProjectKeyInfo interface
func (pk *ProjectKey) IsDisabled() bool {
	return pk.Disabled
}

// RequiresSignature implements ProjectKeyInfo interface
func (pk *ProjectKey) RequiresSignature() bool {
	return pk.SignedOnly
}

// GetSecretEncrypted implements ProjectKeyInfo interface
func (pk *ProjectKey) GetSecretEncrypted() string {
	return pk.SecretEnc
}

// CreateProjectKeyParams contains parameters for creating a project key
type CreateProjectKeyParams struct {
	ProjectID          string
	Label              string
	Env                string
	SignedOnly         bool
	PublicKey          string
	SecretEncrypted    string // AES-GCM encrypted secret
	SecretFingerprint  string // HMAC-SHA256 hashed secret
	CreatedBy          string
}

// ProjectStore is a thin wrapper around Client that provides project-related data access.
// It is used as the concrete implementation injected into handlers.
type ProjectStore struct {
	Client *Client
}

// ProjectKeyStore provides project key-related data access
type ProjectKeyStore struct {
	Client *Client
}

// GetProjectByID => GET /rest/v1/projects?id=eq.<id>
func (s *ProjectStore) GetProjectByID(ctx context.Context, projectID string) (*Project, error) {
	if projectID == "" {
		return nil, errors.New("project id cannot be empty")
	}

	url := s.Client.BaseRestURL + "/projects?id=eq." + projectID + "&select=*"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Supabase REST API headers
	req.Header.Set("apikey", s.Client.ServiceRoleKey)
	req.Header.Set("Authorization", "Bearer "+s.Client.ServiceRoleKey)

	resp, err := s.Client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		log.Printf("supabase rest api error: status=%d body=%s", resp.StatusCode, bodyStr)
		return nil, errors.New(bodyStr)
	}

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, errors.New("project not found")
	}

	return &projects[0], nil
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

// CreateProjectKey => POST /rest/v1/project_keys
func (s *ProjectKeyStore) CreateProjectKey(ctx context.Context, params CreateProjectKeyParams) (*ProjectKey, error) {
	if params.ProjectID == "" {
		return nil, errors.New("project id cannot be empty")
	}
	if params.Label == "" {
		return nil, errors.New("label cannot be empty")
	}
	if params.PublicKey == "" {
		return nil, errors.New("public key cannot be empty")
	}
	if params.SecretEncrypted == "" {
		return nil, errors.New("secret encrypted cannot be empty")
	}
	if params.SecretFingerprint == "" {
		return nil, errors.New("secret fingerprint cannot be empty")
	}

	url := s.Client.BaseRestURL + "/project_keys"

	payload := map[string]interface{}{
		"project_id":         params.ProjectID,
		"label":              params.Label,
		"env":                params.Env,
		"signed_only":        params.SignedOnly,
		"public_key":         params.PublicKey,
		"secret_enc":         params.SecretEncrypted,
		"secret_fingerprint": params.SecretFingerprint,
		"created_by":         params.CreatedBy,
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
		return nil, errors.New(bodyStr)
	}

	var keys []ProjectKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, errors.New("no project key returned from database")
	}

	return &keys[0], nil
}

// GetProjectKeyByPublicKey => GET /rest/v1/project_keys?public_key=eq.<key>
func (s *ProjectKeyStore) GetProjectKeyByPublicKey(ctx context.Context, publicKey string) (*ProjectKey, error) {
	if publicKey == "" {
		return nil, errors.New("public key cannot be empty")
	}

	url := s.Client.BaseRestURL + "/project_keys?public_key=eq." + publicKey + "&select=*"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Supabase REST API headers
	req.Header.Set("apikey", s.Client.ServiceRoleKey)
	req.Header.Set("Authorization", "Bearer "+s.Client.ServiceRoleKey)

	resp, err := s.Client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		log.Printf("supabase rest api error: status=%d body=%s", resp.StatusCode, bodyStr)
		return nil, errors.New(bodyStr)
	}

	var keys []ProjectKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, errors.New("project key not found")
	}

	return &keys[0], nil
}
