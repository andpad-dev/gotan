package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /reports/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "latest")
	})
	mux.HandleFunc("GET /reports/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "report=%s", r.PathValue("id"))
	})

	for _, path := range []string{"/reports/latest", "/reports/2026-08"} {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		fmt.Printf("GET %s -> %d %q\n", path, recorder.Code, recorder.Body.String())
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/reports/2026-08", nil))
	fmt.Printf("POST /reports/2026-08 -> %d\n", recorder.Code)
}
