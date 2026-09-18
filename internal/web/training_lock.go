package web

import "sync"

// Serialize only requests for the same user. References include waiters so an
// entry cannot disappear while a request is waiting for its predecessor.
type trainingUserLocks struct {
	mu      sync.Mutex
	entries map[int64]*trainingUserLock
}
type trainingUserLock struct {
	mu   sync.Mutex
	refs int
}

func (l *trainingUserLocks) lock(userID int64) func() {
	l.mu.Lock()
	if l.entries == nil {
		l.entries = make(map[int64]*trainingUserLock)
	}
	entry := l.entries[userID]
	if entry == nil {
		entry = &trainingUserLock{}
		l.entries[userID] = entry
	}
	entry.refs++
	l.mu.Unlock()
	entry.mu.Lock()
	return func() {
		entry.mu.Unlock()
		l.mu.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(l.entries, userID)
		}
		l.mu.Unlock()
	}
}

func (h *WebTrainingHandler) session(userID int64) (*WebTrainingState, bool) {
	h.sessionsMutex.RLock()
	defer h.sessionsMutex.RUnlock()
	state, ok := h.sessions[userID]
	return state, ok
}
func (h *WebTrainingHandler) setSession(userID int64, state *WebTrainingState) {
	h.sessionsMutex.Lock()
	defer h.sessionsMutex.Unlock()
	h.sessions[userID] = state
}
func (h *WebTrainingHandler) deleteSession(userID int64) {
	h.sessionsMutex.Lock()
	defer h.sessionsMutex.Unlock()
	delete(h.sessions, userID)
}
