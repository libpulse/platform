package handlers

import (
	"context"

	"github.com/libpulse/platform/services/api/internal/supabase"
)

// UserStore abstracts user data access for handlers, enabling dependency injection and unit testing.
type UserStore interface {
	GetUserByID(ctx context.Context, id string) (*supabase.User, error)
}
