package storage

import (
	"sync"
)

type Store struct {
	data map[string]string
	mu   sync.RWMutex
}


// New creates a new empty store.
// New يسوي مخزن فارغ جديد.
func New() *Store {
	return &Store{
		data: make(map[string]string),
		mu:   sync.RWMutex{},
	}
}

// Set Stores a key-vlaue pair in the store.
// Set يحفظ مفتاح-قيمة معينة في المخزن.
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get Retrieves the value associated with the given key from the store.
// Get يبحث عن قيمة مفتاح معينة في المخزن.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

// Delete removes the key-value pairs from the store.
// Returns the number of keys that were deleted.
// Delete يحذف مفاتيح-قيم معينة من المخزن.
// يرجع عدد المفاتيح اللي انحذفت.
func (s *Store) Delete(keys []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, key := range keys {
		if _, exists := s.data[key]; exists {
			delete(s.data, key)
			count++
		}
	}
	return count
}

// Exists returns the number of keys that exist in the store.
// Exists يرجع عدد المفاتيح الموجودة في المخزن.
func (s *Store) Exists(keys []string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, key := range keys {
		if _, exists := s.data[key]; exists {
			count++
		}
	}
	return count
}



// Keys returns all keys matching a pattern.
// For now, "*" returns all keys (we'll add pattern matching later).
//
// Keys يرجع كل المفاتيح المطابقة لـ pattern.
// الحين، "*" يرجع كل المفاتيح.
func (s *Store) Keys(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if pattern == "*" {
		// Return all keys
		// ارجع كل المفاتيح
		keys := make([]string, 0, len(s.data))
		for k := range s.data {
			keys = append(keys, k)
		}
		return keys
	}

	// For now, only support "*"
	// الحين، دعم بس "*"
	return []string{}
}