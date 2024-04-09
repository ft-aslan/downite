package api

import (
	"downite/handlers"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
)

// Options for the CLI.
type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"9999"`
}

func ApiInit() {
	// Create a CLI app which takes a port option.
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		s := http.NewServeMux()

		//initilize docs
		s.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
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
	      data-url="/api/openapi.json"></script>
	    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	  </body>
	</html>`))
		})

		//initilize huma
		config := huma.DefaultConfig("Downite API", "0.0.1")
		config.Servers = []*huma.Server{{URL: "http://localhost:9999/api"}}

		config.OpenAPIPath = "/openapi"
		config.DocsPath = ""

		api := humago.NewWithPrefix(s, "/api", config)

		//register api routes
		huma.Register(api, huma.Operation{
			OperationID: "get-all-torrents",
			Method:      http.MethodGet,
			Path:        "/torrent",
			Summary:     "Get all torrents",
		}, handlers.GetTorrents)
		huma.Post(api, "/torrent", handlers.DownloadTorrent)
		huma.Get(api, "/torrent/:hash", handlers.GetTorrent)
		huma.Post(api, "/torrent-meta", handlers.GetTorrentMeta)

		//write api json to file
		apiJson, err := json.Marshal(api.OpenAPI())
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("docs/openapi.json", apiJson, 0644)
		if err != nil {
			fmt.Println("Error writing openapi to file:", err)
			return
		}
		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), s)
		})
	})
	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
