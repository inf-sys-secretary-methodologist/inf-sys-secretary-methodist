package scenario

import "sync"

// SharedState holds state shared between scenario steps.
type SharedState struct {
	mu            sync.RWMutex
	CreatedDocs   map[string]int64
	CreatedTasks  map[string]int64
	CreatedEvents map[string]int64
	Conversations map[string]int64
	Reports       map[string]int64
	Extra         map[string]any
}

// NewSharedState creates a new empty shared state.
func NewSharedState() *SharedState {
	return &SharedState{
		CreatedDocs:   make(map[string]int64),
		CreatedTasks:  make(map[string]int64),
		CreatedEvents: make(map[string]int64),
		Conversations: make(map[string]int64),
		Reports:       make(map[string]int64),
		Extra:         make(map[string]any),
	}
}

// SetDoc stores a document ID.
func (s *SharedState) SetDoc(key string, id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CreatedDocs[key] = id
}

// GetDoc retrieves a stored document ID.
func (s *SharedState) GetDoc(key string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.CreatedDocs[key]
	return id, ok
}

// SetTask stores a task ID.
func (s *SharedState) SetTask(key string, id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CreatedTasks[key] = id
}

// GetTask retrieves a stored task ID.
func (s *SharedState) GetTask(key string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.CreatedTasks[key]
	return id, ok
}

// SetEvent stores an event ID.
func (s *SharedState) SetEvent(key string, id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CreatedEvents[key] = id
}

// GetEvent retrieves a stored event ID.
func (s *SharedState) GetEvent(key string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.CreatedEvents[key]
	return id, ok
}

// SetConversation stores a conversation ID.
func (s *SharedState) SetConversation(key string, id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Conversations[key] = id
}

// GetConversation retrieves a stored conversation ID.
func (s *SharedState) GetConversation(key string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.Conversations[key]
	return id, ok
}

// SetReport stores a report ID.
func (s *SharedState) SetReport(key string, id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Reports[key] = id
}

// GetReport retrieves a stored report ID.
func (s *SharedState) GetReport(key string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.Reports[key]
	return id, ok
}

// SetExtra stores an arbitrary value.
func (s *SharedState) SetExtra(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Extra[key] = val
}

// GetExtra retrieves an arbitrary value.
func (s *SharedState) GetExtra(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.Extra[key]
	return val, ok
}
