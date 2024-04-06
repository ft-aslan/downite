package api

import (
	"downite/handlers"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func ApiInit() error {
	fmt.Println("Starting Downite server...")
	config := huma.DefaultConfig("Downite API", "0.0.1")
	config.OpenAPIPath = "/doc/openapi"
	config.DocsPath = ""
	s := http.NewServeMux()

	s.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!doctype html>
	<html>
	  <head>
	    <title>API Reference</title>
	    <meta charset="utf-8" />
	    <meta
	      name="viewport"
	      content="width=device-width, initial-scale=1" />
	  </head>
	  <body>
	    <script
	      id="api-reference"
	      data-url="/openapi.json"></script>
	    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	  </body>
	</html>`))
	})

	api := humago.NewWithPrefix(s, "/api", config)

	huma.Register(api, huma.Operation{
		OperationID: "get-all-torrents",
		Method:      http.MethodGet,
		Path:        "/torrent",
		Summary:     "Get all torrents",
	}, handlers.GetTorrents)
	// huma.Post(api, "/torrent", handlers.DownloadTorrent)
	// huma.Get(api, "/torrent/:hash", handlers.GetTorrent)
	// huma.Post(api, "/torrent-meta", handlers.GetTorrentMeta)

	// s := fuego.NewServer(fuego.WithCorsMiddleware(cors.New(cors.Options{
	// 	AllowedOrigins: []string{"*"},
	// 	AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	// }).Handler))

	err := http.ListenAndServe(":9999", s)
	return err
}
