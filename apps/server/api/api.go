package api

import (
	"downite/handlers"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/rs/cors"
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
			w.Write([]byte(`
			<!doctype html>
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
			</html>
			`))
		})

		//initilize huma
		config := huma.DefaultConfig("Downite API", "0.0.1")
		config.Servers = []*huma.Server{{URL: "http://localhost:9999/api"}}

		config.OpenAPIPath = "/openapi"
		config.DocsPath = ""

		api := humago.NewWithPrefix(s, "/api", config)
		// api.UseMiddleware(CorsMiddleware)

		//register torrent routes
		AddTorrentRoutes(api)
		//register download routes
		AddDownloadRoutes(api)

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

		//run prettier for openapi.json
		err = exec.Command("bunx", "prettier", "docs/openapi.json", "--write", "--parser", "json").Run()
		if err != nil {
			fmt.Println("Error running prettier for openapi.json:", err)
			return
		}

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)

			//disabled cors
			s := cors.New(cors.Options{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			}).Handler(s)

			http.ListenAndServe(fmt.Sprintf("localhost:%d", options.Port), s)
		})
	})
	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
func AddTorrentRoutes(api huma.API) {
	//register api routes
	// registering the download torrent route manually because it's a multipart/form-data request
	schema := api.OpenAPI().Components.Schemas.Schema(reflect.TypeOf(handlers.DownloadTorrentReqBody{}), true, "DownloadTorrentReqBodyStruct")
	huma.Register(api, huma.Operation{
		OperationID: "download-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent",
		Summary:     "Download torrent",
		RequestBody: &huma.RequestBody{
			Content: map[string]*huma.MediaType{
				"multipart/form-data": {
					Schema: schema,
					Encoding: map[string]*huma.Encoding{
						"torrentFile": {
							ContentType: "application/x-bittorrent",
						},
					},
				},
			},
		},
	}, handlers.DownloadTorrent)
	huma.Register(api, huma.Operation{
		OperationID: "get-all-torrents",
		Method:      http.MethodGet,
		Path:        "/torrent",
		Summary:     "Get all torrents",
	}, handlers.GetTorrents)
	huma.Register(api, huma.Operation{
		OperationID: "get-torrent",
		Method:      http.MethodGet,
		Path:        "/torrent/{infohash}",
		Summary:     "Get torrent",
	}, handlers.GetTorrent)
	huma.Register(api, huma.Operation{
		OperationID: "pause-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent/pause",
		Summary:     "Pause torrent",
	}, handlers.PauseTorrent)
	huma.Register(api, huma.Operation{
		OperationID: "resume-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent/resume",
		Summary:     "Resume torrent",
	}, handlers.ResumeTorrent)
	huma.Register(api, huma.Operation{
		OperationID: "remove-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent/remove",
		Summary:     "Remove torrent",
	}, handlers.RemoveTorrent)
	huma.Register(api, huma.Operation{
		OperationID: "delete-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent/delete",
		Summary:     "Delete torrent",
	}, handlers.DeleteTorrent)
	huma.Register(api, huma.Operation{
		OperationID: "get-torrent-meta-info-with-magnet",
		Method:      http.MethodPost,
		Path:        "/meta/magnet",
		Summary:     "Get torrent meta info with magnet",
	}, handlers.GetMetaWithMagnet)
	huma.Register(api, huma.Operation{
		OperationID: "get-torrent-meta-info-with-file",
		Method:      http.MethodPost,
		Path:        "/meta/file",
		Summary:     "Get torrent meta info with file",
	}, handlers.GetMetaWithFile)
	huma.Register(api, huma.Operation{
		OperationID: "get-torrents-total-speed",
		Method:      http.MethodGet,
		Path:        "/torrent/speed",
		Summary:     "Get torrents total speed",
	}, handlers.GetTorrentsTotalSpeed)

}
func AddDownloadRoutes(api huma.API) {
}

// Create a custom middleware handler to disable CORS
// func CorsMiddleware(ctx huma.Context, next func(huma.Context)) {
// 	ctx.SetHeader("Access-Control-Allow-Origin", "*")
// 	ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 	ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

// 	// Call the next middleware in the chain. This eventually calls the
// 	// operation handler as well.
// 	next(ctx)
// }
