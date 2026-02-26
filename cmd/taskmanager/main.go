package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxBodyBytes = 1024 * 1024

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(data); err != nil {
		var syntaxErr *json.SyntaxError
		var maxErr *http.MaxBytesError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("syntax error in json %w", err)
		case errors.As(err, &maxErr):
			return fmt.Errorf("request body is too large (maximum 1MB)")
		default:
			return fmt.Errorf("failed to decode JSON: %w", err)
		}
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("request body contains invalid JSON: too many objects")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("writeJSON encode error: %v\n", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{"status": "ok"}
		writeJSON(w, http.StatusOK, data)
	})

	mux.HandleFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		var data json.RawMessage
		if err := readJSON(w, r, &data); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, &data)
	})

	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "page not found")
	})

	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	fmt.Printf("starting server at %s\n", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		fmt.Println("failed to start server")
		fmt.Println(err)
	}
}
