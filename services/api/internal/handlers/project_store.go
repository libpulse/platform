package handlers

import (
	"context"

	"github.com/libpulse/platform/services/api/internal/supabase"
)

// ProjectStore abstracts project data access for handlers, enabling dependency injection and unit testing.
type ProjectStore interface {
	GetProjectByID(ctx context.Context, projectID string) (*supabase.Project, error)
	CreateProject(ctx context.Context, name string, ownerUserID string) (*supabase.Project, error)
}

// ProjectKeyStore abstracts project key data access for handlers, enabling dependency injection and unit testing.
type ProjectKeyStore interface {
	CreateProjectKey(ctx context.Context, params supabase.CreateProjectKeyParams) (*supabase.ProjectKey, error)
}
