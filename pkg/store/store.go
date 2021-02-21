package store

import (
	"errors"
)

var ErrorKeyNotFound = errors.New("Key not found")

var store = make(map[string]string)

func Put(k string, v string) error {
	store[k] = v
	return nil
}

func Get(k string) (string, error) {
	v, ok := store[k]
	if !ok {
		return "", ErrorKeyNotFound
	}

	return v, nil
}

func Delete(k string) error {
	delete(store, k)
	return nil
}
