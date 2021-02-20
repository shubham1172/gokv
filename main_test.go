package main

import (
	"testing"
)

func TestPut(t *testing.T) {
	testCases := []struct {
		pkey string // key for Put
		pval string // value for Put
		gval string // expected value from Get
	}{
		{"testPutKey1", "value1", "value1"},
		{"testPutKey1", "value2", "value2"},
		{"testPutKey2", "", ""},
	}

	for _, tc := range testCases {
		Put(tc.pkey, tc.pval)
		v, _ := Get(tc.pkey)

		if v != tc.gval {
			t.Errorf("Value was incorrect, expected: %s, got: %s", tc.gval, v)
		}
	}
}

func TestGet(t *testing.T) {
	// missing value
	v, err := Get("testGetKey1")
	if err != ErrorKeyNotFound {
		t.Errorf("Expected to throw %v, got %v, value: %s", ErrorKeyNotFound, err, v)
	}

	// value present in store
	v0 := "value1"
	Put("testGetKey1", v0)
	v, err = Get("testGetKey1")

	if v != v0 {
		t.Errorf("Value was incorrect, expected %s, got %s", v0, v)
	}

	if err != nil {
		t.Errorf("Expected err to be nil, got %v instead", err)
	}
}

func TestDelete(t *testing.T) {
	Put("testKeyDelete1", "value1")

	testCases := []struct {
		val string
		err error
	}{
		{"testDeleteKey1", nil}, // value already in store
		{"testDeleteKey2", nil}, // value not in store
	}

	for _, tc := range testCases {
		err := Delete(tc.val)
		if err != nil {
			t.Errorf("Expected err to be %v, got %v instead, value: %s", tc.err, err, tc.val)
		}
	}
}
