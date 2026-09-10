package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /x/fixed", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "fixed")
	})
	mux.HandleFunc("GET /x/{value}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "value=%s", r.PathValue("value"))
	})

	for _, path := range []string{"/x/fixed", "/x/other"} {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		fmt.Printf("GET %s -> %d %q\n", path, recorder.Code, recorder.Body.String())
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/x/other", nil))
	fmt.Printf("POST /x/other -> %d\n", recorder.Code)
}
