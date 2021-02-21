package server

import (
	"github.com/gorilla/mux"
	"github.com/shubham1172/gokv/pkg/store"
	"io/ioutil"
	"log"
	"net/http"
)

const messageKeyNotFound string = "Key missing. Usage: /api/v1/key/:key"
const messageValueNotFound string = "Value missing in the request body"

// serves PUT /api/v1/key/{key}
func keyPutHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, messageKeyNotFound, http.StatusBadRequest)
		return
	}

	value, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(value) == 0 {
		http.Error(w, messageValueNotFound, http.StatusBadRequest)
		return
	}

	err = store.Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// serves GET /api/v1/key/{key}
func keyGetHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, messageKeyNotFound, http.StatusBadRequest)
		return
	}

	value, err := store.Get(key)
	if err != nil {
		if err == store.ErrorKeyNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Write([]byte(value))
}

/// serves DELETE /api/v1/key/{key}
func keyDeleteHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, messageKeyNotFound, http.StatusBadRequest)
		return
	}

	err := store.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Start the http server on the given address.
func Start(addr string) {
	r := mux.NewRouter()

	// register routes
	r.HandleFunc("/api/v1/key/{key}", keyPutHandler).Methods("PUT")
	r.HandleFunc("/api/v1/key/{key}", keyGetHandler).Methods("GET")
	r.HandleFunc("/api/v1/key/{key}", keyDeleteHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(addr, r))
}
