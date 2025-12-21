package handlers

import (
	"context"

	"github.com/libpulse/platform/services/api/internal/supabase"
)

// ProjectStore abstracts project data access for handlers, enabling dependency injection and unit testing.
type ProjectStore interface {
	CreateProject(ctx context.Context, name string, ownerUserID string) (*supabase.Project, error)
}
