package store

import (
	"strings"
	"testing"
)

func getALongString() string {
	return strings.Repeat("a", 1025)
}

func TestPut(t *testing.T) {
	testCases := []struct {
		name string
		pkey string // key for Put
		pval string // value for Put
		gval string // expected value from Get
		err  error  // expected error from Put
	}{
		{"put new key-value", "testPutKey1", "value1", "value1", nil},
		{"overwrite existing key", "testPutKey1", "value2", "value2", nil},
		{"put empty value", "testPutKey2", "", "", nil},
		{"really long key", getALongString(), "value3", "", ErrorKeySizeTooLarge},
		{"really long value", "testPutKey3", getALongString(), "", ErrorValueSizeTooLarge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Put(tc.pkey, tc.pval)
			if err != tc.err {
				t.Errorf("Error was incorrect, expected: %v, got: %v", tc.err, err)
			}

			// only try a get if there was no error during put
			if err != nil {
				return
			}

			v, _ := Get(tc.pkey)
			if v != tc.gval {
				t.Errorf("Value was incorrect, expected: %s, got: %s", tc.gval, v)
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Run("missing value", func(t *testing.T) {
		v, err := Get("testGetKey1")
		if err != ErrorKeyNotFound {
			t.Errorf("Expected to throw %v, got %v, value: %s", ErrorKeyNotFound, err, v)
		}
	})

	t.Run("value present in store", func(t *testing.T) {
		v0 := "value1"
		Put("testGetKey1", v0)
		v, err := Get("testGetKey1")

		if v != v0 {
			t.Errorf("Value was incorrect, expected %s, got %s", v0, v)
		}

		if err != nil {
			t.Errorf("Expected err to be nil, got %v instead", err)
		}
	})

	t.Run("query key too large", func(t *testing.T) {
		_, err := Get(getALongString())

		if err != ErrorKeySizeTooLarge {
			t.Errorf("Expected err to be nil, got %v instead", err)
		}
	})
}

func TestDelete(t *testing.T) {
	Put("testKeyDelete1", "value1")

	testCases := []struct {
		name string
		val  string
		err  error
	}{
		{"value already in store", "testDeleteKey1", nil},
		{"value not in store", "testDeleteKey2", nil},
		{"key too large", getALongString(), ErrorKeySizeTooLarge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Delete(tc.val)
			if err != tc.err {
				t.Errorf("Expected err to be %v, got %v instead, value: %s", tc.err, err, tc.val)
			}
		})
	}
}
