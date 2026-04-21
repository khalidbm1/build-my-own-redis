package storage

import (
	"sync"
	"time"
)

type DataType int

const (
	TypeString DataType = iota
	TypeList
	TypeSet
)

type Entry struct {
	Type     DataType
	Value    interface{}
	ExpireAt *time.Time
}

type Store struct {
	data map[string]*Entry
	mu   sync.RWMutex
}

// New creates a new empty store.
// New يسوي مخزن فارغ جديد.
func New() *Store {
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

	if s.isExpired(entry) {
		return "", false

	}
	if entry.Type != TypeString {
		return "", false
	}
	return entry.Value.(string), true
}

func (s *Store) LPush(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		entry = &Entry{Type: TypeList, Value: []string{}}
		s.data[key] = entry
	}

	if entry.Type != TypeList {
		return -1 // Wrong type
	}

	list := entry.Value.([]string)
	// Prepend all values
	// أضف كل القيم في البداية
	for i := len(values) - 1; i >= 0; i-- {
		list = append([]string{values[i]}, list...)
	}
	entry.Value = list
	return len(list)
}

func (s *Store) RPush(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		entry = &Entry{Type: TypeList, Value: []string{}}
		s.data[key] = entry
	}

	if entry.Type != TypeList {
		return -1
	}

	list := entry.Value.([]string)
	list = append(list, values...)
	entry.Value = list
	return len(list)
}

func (s *Store) LPop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		return "", false
	}

	if entry.Type != TypeList {
		return "", false
	}

	list := entry.Value.([]string)
	if len(list) == 0 {
		return "", false
	}

	value := list[0]
	entry.Value = list[1:]
	return value, true
}

func (s *Store) RPop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		return "", false
	}

	if entry.Type != TypeList {
		return "", false
	}

	list := entry.Value.([]string)
	if len(list) == 0 {
		return "", false
	}

	value := list[len(list)-1]
	entry.Value = list[:len(list)-1]
	return value, true
}

func (s *Store) LLen(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return 0
	}

	if entry.Type != TypeList {
		return 0
	}

	return len(entry.Value.([]string))
}

func (s *Store) LRange(key string, start, stop int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return []string{}
	}

	if entry.Type != TypeList {
		return []string{}
	}

	list := entry.Value.([]string)
	length := len(list)

	// Handle negative indices
	// تعامل مع الفهارس السالبة
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []string{}
	}

	return list[start : stop+1]
}

// Set operations (stored as map[string]bool)
func (s *Store) SAdd(key string, members ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		entry = &Entry{Type: TypeSet, Value: make(map[string]bool)}
		s.data[key] = entry
	}

	if entry.Type != TypeSet {
		return -1
	}

	set := entry.Value.(map[string]bool)
	added := 0
	for _, member := range members {
		if !set[member] {
			set[member] = true
			added++
		}
	}
	return added
}

func (s *Store) SRem(key string, members ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		return 0
	}

	if entry.Type != TypeSet {
		return 0
	}

	set := entry.Value.(map[string]bool)
	removed := 0
	for _, member := range members {
		if set[member] {
			delete(set, member)
			removed++
		}
	}
	return removed
}

func (s *Store) SMembers(key string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return []string{}
	}

	if entry.Type != TypeSet {
		return []string{}
	}

	set := entry.Value.(map[string]bool)
	members := make([]string, 0, len(set))
	for member := range set {
		members = append(members, member)
	}
	return members
}

func (s *Store) SIsMember(key, member string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return false
	}

	if entry.Type != TypeSet {
		return false
	}

	set := entry.Value.(map[string]bool)
	return set[member]
}

func (s *Store) SCard(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return 0
	}

	if entry.Type != TypeSet {
		return 0
	}

	return len(entry.Value.(map[string]bool))
}

// Helper method
func (s *Store) isExpired(entry *Entry) bool {
	if entry.ExpireAt == nil {
		return false
	}
	return time.Now().After(*entry.ExpireAt)
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
