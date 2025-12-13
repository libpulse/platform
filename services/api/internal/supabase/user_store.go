package supabase

import "context"

// UserStore is a thin wrapper around Client that provides user-related data access.
// It is used as the concrete implementation injected into handlers.
type UserStore struct {
	Client *Client
}

// GetUserByID delegates to the underlying Supabase client.
func (s *UserStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.Client.GetUserByID(ctx, id)
}
