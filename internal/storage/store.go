package storage

import (
	"sync"
	"time"
)

type Entry struct {
	Value    string
	ExpireAt *time.Time
}

type Store struct {
	data map[string]*Entry
	mu   sync.RWMutex
}

// New creates a new empty store.
// New يسوي مخزن فارغ جديد.
func New() *Store {
	// return &Store{
	// 	data: make(map[string]string),
	// 	mu:   sync.RWMutex{},
	// }
	s := &Store{
		data: make(map[string]*Entry),
	}
	go s.cleanExpiredKeys()
	return s
}

// Set Stores a key-vlaue pair in the store.
// Set يحفظ مفتاح-قيمة معينة في المخزن.
func (s *Store) Set(key, value string, ttl *time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := &Entry{Value: value}

	if ttl != nil {
		expireAt := time.Now().Add(*ttl)
		entry.ExpireAt = &expireAt
	}
	s.data[key] = entry
}

// Get Retrieves the value associated with the given key from the store.
// Get يبحث عن قيمة مفتاح معينة في المخزن.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	//value, exists := s.data[key]
	entry, exists := s.data[key]

	if !exists {
		return "", false
	}

	if entry.ExpireAt != nil && time.Now().After(*entry.ExpireAt) {
		return "", false

	}
	return entry.Value, true
}

func (s *Store) Expire(key string, ttl *time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.data[key]

	if !exists {
		return false
	}

	if ttl == nil {
		entry.ExpireAt = nil
	} else {
		expireAt := time.Now().Add(*ttl)
		entry.ExpireAt = &expireAt
	}
	return true
}

func (s *Store) TTL(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, exists := s.data[key]

	if !exists {
		return -2
	}

	if entry.ExpireAt == nil {
		return -1
	}

	remaining := time.Until(*entry.ExpireAt).Seconds()

	if remaining <= 0 {
		return -2
	}

	return int(remaining)
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

func (s *Store) cleanExpiredKeys() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()

		for key, entry := range s.data {
			if entry.ExpireAt != nil && now.After(*entry.ExpireAt) {
				delete(s.data, key)
			}
		}
		s.mu.Unlock()
	}
}
