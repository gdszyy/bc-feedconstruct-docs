// Package main is the entrypoint for the BFF service.
//
// BDD placeholder — actual wiring is implemented after BDD test files
// are confirmed by the user (see CLAUDE.md / docs/08_backend_railway/).
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintln(w, "not-ready: bff in BDD scaffold phase")
	})

	fmt.Fprintf(os.Stdout, "bffd listening on :%s (BDD scaffold; no business logic yet)\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "bffd exited: %v\n", err)
		os.Exit(1)
	}
}
