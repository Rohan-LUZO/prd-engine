package auth

// MemoryUserStore is an in-memory UserStore for tests.
type MemoryUserStore struct {
	UsersByToken map[string]*User
}

// NewMemoryUserStore returns a store with the given token -> user map.
func NewMemoryUserStore(usersByToken map[string]*User) *MemoryUserStore {
	if usersByToken == nil {
		usersByToken = make(map[string]*User)
	}
	return &MemoryUserStore{UsersByToken: usersByToken}
}

// FindByToken implements UserStore.
func (s *MemoryUserStore) FindByToken(token string) (*User, error) {
	if u, ok := s.UsersByToken[token]; ok {
		return u, nil
	}
	return nil, nil
}
