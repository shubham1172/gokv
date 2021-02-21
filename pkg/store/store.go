package store

import (
	"errors"
	"fmt"
	"sync"
)

const (
	maxKeySize   = 1024
	maxValueSize = 1024
)

var (
	// ErrorKeyNotFound is returned by the Get method to indicate that the key was not present in the store.
	ErrorKeyNotFound = errors.New("Key not found")

	// ErrorKeySizeTooLarge is returned to indicate that the key size is more than the max permittable size.
	ErrorKeySizeTooLarge = fmt.Errorf("Key size too large, max permissible: %d", maxKeySize)

	// ErrorValueSizeTooLarge is return to indicate that the value size is more than the max permittable size.
	ErrorValueSizeTooLarge = fmt.Errorf("Value size too large, max permissible: %d", maxValueSize)
)

var store = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

// Put a value in the store against a key. If the key already exists,
// it is overwritten.
func Put(k string, v string) error {
	if len(k) > maxKeySize {
		return ErrorKeySizeTooLarge
	}
	if len(v) > maxValueSize {
		return ErrorValueSizeTooLarge
	}

	store.Lock()
	store.m[k] = v
	store.Unlock()

	return nil
}

// Get returns a value from the store associated with a key.
// Returns ErrorKeyNotFound if key does not exist.
func Get(k string) (string, error) {
	if len(k) > maxKeySize {
		return "", ErrorKeySizeTooLarge
	}

	store.RLock()
	v, ok := store.m[k]
	store.RUnlock()

	if !ok {
		return "", ErrorKeyNotFound
	}

	return v, nil
}

// Delete ensures that a key does not exist in the store.
// If a key is missing, the function passes silently.
func Delete(k string) error {
	if len(k) > maxKeySize {
		return ErrorKeySizeTooLarge
	}

	delete(store.m, k)
	return nil
}
