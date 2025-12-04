package supabase

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// User structure response externally (aligned with openapi.yaml).
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      *string   `json:"name,omitempty"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Supabase Admin API UserResponse
type supabaseUserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AppMetadata struct {
		Provider string `json:"provider"`
	} `json:"app_metadata"`
	UserMetadata struct {
		FullName  string `json:"full_name"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user_metadata"`
}

// Client Wrapper
type Client struct {
	BaseAuthURL    string
	ServiceRoleKey string
	httpClient     *http.Client
}

// authBaseURL. e.g. https://<project-id>.supabase.co/auth/v1
func NewClient(authBaseURL, serviceRoleKey string) *Client {
	return &Client{
		BaseAuthURL:    strings.TrimRight(authBaseURL, "/"),
		ServiceRoleKey: serviceRoleKey,
		httpClient:     &http.Client{Timeout: 5 * time.Second},
	}
}

// GetUserByID => GET /auth/v1/admin/users/{id}
func (c *Client) GetUserByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, errors.New("empty user id")
	}

	url := c.BaseAuthURL + "/admin/users/" + id

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Supabase Admin API needs apikey + Authorization: Bearer service_role
	req.Header.Set("apikey", c.ServiceRoleKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceRoleKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("user not found")
	}

	// Print the error if status code is over 400.
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("supabase admin error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
		return nil, errors.New("supabase admin api error: " + resp.Status)
	}

	var su supabaseUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&su); err != nil {
		return nil, err
	}

	u := &User{
		ID:        su.ID,
		Email:     su.Email,
		Provider:  su.AppMetadata.Provider,
		CreatedAt: su.CreatedAt,
		UpdatedAt: su.UpdatedAt,
	}

	if su.UserMetadata.FullName != "" {
		name := su.UserMetadata.FullName
		u.Name = &name
	}
	if su.UserMetadata.AvatarURL != "" {
		avatar := su.UserMetadata.AvatarURL
		u.AvatarURL = &avatar
	}

	return u, nil
}
