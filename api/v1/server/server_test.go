package server

import (
	"github.com/gorilla/mux"
	"github.com/shubham1172/gokv/internal/logger"
	"github.com/shubham1172/gokv/pkg/store"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type dummyLogger struct{}

func (d *dummyLogger) WriteDelete(key string)                          {}
func (d *dummyLogger) WritePut(key, value string)                      {}
func (d *dummyLogger) Err() <-chan error                               { return nil }
func (d *dummyLogger) ReadEvents() (<-chan logger.Event, <-chan error) { return nil, nil }
func (d *dummyLogger) Run()                                            {}
func (d *dummyLogger) Stop()                                           {}

func getALongString() string {
	return strings.Repeat("a", 1025)
}

func TestKeyPutHandler(t *testing.T) {
	testCases := []struct {
		name       string
		key        string
		value      string
		statusCode int
	}{
		{"missing key", "", "testKeyPutHandlerValue1", http.StatusBadRequest},
		{"missing value", "testKeyPutHandlerKey1", "", http.StatusBadRequest},
		{"valid request", "testKeyPutHandlerKey1", "testKeyPutHandlerValue1", http.StatusCreated},
		{"really long key", getALongString(), "testKeyPutHandlerValue2", http.StatusBadRequest},
		{"really long value", "testKeyPutHandlerKey2", getALongString(), http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("PUT", "localhost:8080/api/v1/key/"+tc.key, strings.NewReader(tc.value))
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			req = mux.SetURLVars(req, map[string]string{
				"key": tc.key,
			})

			rec := httptest.NewRecorder()
			keyPutHandler(rec, req, &dummyLogger{})

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tc.statusCode {
				t.Errorf("expected status %d, got %d instead", tc.statusCode, res.StatusCode)
			}
		})
	}
}

func TestKeyGetHandler(t *testing.T) {
	testCases := []struct {
		name       string
		key        string
		statusCode int
		resp       string
	}{
		{"missing key (URL)", "", http.StatusBadRequest, ""},
		{"missing key (store)", "testKeyGetHandlerKey1", http.StatusNotFound, ""},
		{"valid key", "testKeyGetHandlerKey2", http.StatusOK, "testKeyGetHandlerValue2"},
		{"really long key", getALongString(), http.StatusBadRequest, ""},
	}

	store.Put("testKeyGetHandlerKey2", "testKeyGetHandlerValue2")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "localhost:8080/api/v1/key/"+tc.key, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			req = mux.SetURLVars(req, map[string]string{
				"key": tc.key,
			})

			rec := httptest.NewRecorder()
			keyGetHandler(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			if res.StatusCode != tc.statusCode {
				t.Errorf("expected status %d, got %d instead", tc.statusCode, res.StatusCode)
			}
			if tc.resp != "" && string(b) != tc.resp {
				t.Errorf("expected response %s, got %s instead", tc.resp, string(b))
			}
		})
	}
}

func TestKeyDeleteHandler(t *testing.T) {
	testCases := []struct {
		name       string
		key        string
		statusCode int
	}{
		{"missing key", "", http.StatusBadRequest},
		{"missing key (store)", "testKeyDeleteHandlerKey1", http.StatusOK},
		{"valid key", "testKeyDeleteHandlerKey2", http.StatusOK},
		{"really long key", getALongString(), http.StatusBadRequest},
	}

	store.Put("testKeyDeleteHandlerKey2", "testKeyDeleteHandlerValue2")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "localhost:8080/api/v1/key/"+tc.key, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			req = mux.SetURLVars(req, map[string]string{
				"key": tc.key,
			})

			rec := httptest.NewRecorder()
			keyDeleteHandler(rec, req, &dummyLogger{})

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.statusCode {
				t.Errorf("expected status %d, got %d instead", tc.statusCode, res.StatusCode)
			}
		})
	}
}
