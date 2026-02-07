package slogx

// RemovalSet is a user-defined collection of attribute keys that should be
// removed from log output. It mirrors the API style of MaskRules.
type RemovalSet struct {
	keys []string
}

// NewRemovalSet creates an empty RemovalSet or initializes it with keys.
func NewRemovalSet(keys ...string) *RemovalSet {
	return &RemovalSet{keys: keys}
}

// Add appends one or more keys to the removal set.
func (s *RemovalSet) Add(keys ...string) *RemovalSet {
	s.keys = append(s.keys, keys...)
	return s
}

// Keys returns the underlying slice of keys.
func (s *RemovalSet) Keys() []string {
	return s.keys
}
