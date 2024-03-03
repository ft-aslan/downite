package api

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
)

func SetupApi() {
	fmt.Println("Starting Downite server...")
	s := fuego.NewServer()
	mux := http.NewServeMux()

	// Create a sub-router for /api/v1
	apiV1Mux := http.NewServeMux()
	apiV1Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	// Mount the /api/v1 sub-router onto the main ServeMux
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1Mux))
	fmt.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", mux)
}
