package store

import (
	"errors"
)

// ErrorKeyNotFound is returned by the Get method to indicate
// that the key was not present in the store.
var ErrorKeyNotFound = errors.New("Key not found")

var store = make(map[string]string)

// Put a value in the store against a key. If the key already exists,
// it is overwritten.
func Put(k string, v string) error {
	store[k] = v
	return nil
}

// Get returns a value from the store associated with a key.
// Returns ErrorKeyNotFound if key does not exist.
func Get(k string) (string, error) {
	v, ok := store[k]
	if !ok {
		return "", ErrorKeyNotFound
	}

	return v, nil
}

// Delete ensures that a key does not exist in the store.
// If a key is missing, the function passes silently.
func Delete(k string) error {
	delete(store, k)
	return nil
}
