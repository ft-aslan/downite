package api

import (
	"context"
	"downite/db"
	"downite/download/protocol/direct"
	"downite/download/protocol/torr"
	"downite/handlers"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/rs/cors"
)

type ApiOptions struct {
	Port int `help:"Port to listen on" short:"p" default:"9999"`
}
type API struct {
	Cli     humacli.CLI
	humaApi huma.API
	Options *ApiOptions
}

func ApiInit(options *ApiOptions) *API {
	api := &API{}
	cli := humacli.New(func(hooks humacli.Hooks, options *ApiOptions) {
		mux := http.NewServeMux()
		//initilize docs
		mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
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

		humaApi := humago.NewWithPrefix(mux, "/api", config)
		api.humaApi = humaApi
		// api.UseMiddleware(CorsMiddleware)

		//disabled cors
		cors := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		})
		corsMux := cors.Handler(mux)

		server := http.Server{
			Addr:    fmt.Sprintf("localhost:%d", options.Port),
			Handler: corsMux,
		}

		//initilize db
		db, err := db.DbInit()
		if err != nil {
			fmt.Printf("Cannot connect to db : %s", err)
		}

		//initilize torrent engine
		torrentEngine, err := InitTorrentEngine(db)
		//initilize download client
		downloadEngine, err := InitDownloadEngine(db)
		//register download routes
		AddDownloadRoutes(handlers.DownloadHandler{
			Db:     db,
			Engine: downloadEngine,
		}, api.humaApi)
		//register torrent routes
		AddTorrentRoutes(handlers.TorrentHandler{
			Db:     db,
			Engine: torrentEngine,
		}, api.humaApi)
		AddSystemRoutes(handlers.SystemHandler{}, api.humaApi)

		api.ExportOpenApi()

		// Tell the CLI how to start your router.
		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", api.Options.Port)
			server.ListenAndServe()
		})

		// Tell the CLI how to stop your server.
		hooks.OnStop(func() {
			errs := torrentEngine.Stop()
			if len(errs) > 0 {
				for _, err := range errs {
					fmt.Printf("Error while stopping torrent engine : %s", err)
				}
			}
			errs = downloadEngine.Stop()
			if len(errs) > 0 {
				for _, err := range errs {
					fmt.Printf("Error while stopping download client : %s", err)
				}
			}
			fmt.Printf("Stopping server...\n")
			// Give the server 5 seconds to gracefully shut down, then give up.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		})
	})

	api.Options = options
	api.Cli = cli

	return api
}
func (api *API) Run() {
	api.Cli.Run()
}
func InitTorrentEngine(db *db.Database) (*torr.TorrentEngine, error) {
	//initilize torrent engine
	pieceCompletionDir := "./tmp"
	defaultTorrentsDir := "./tmp/torrents"
	torrentEngineConfig := torr.TorrentEngineConfig{
		PieceCompletionDbPath: pieceCompletionDir,
		DownloadPath:          defaultTorrentsDir,
	}
	torrentEngine, err := torr.CreateTorrentEngine(torrentEngineConfig, db)
	if err != nil {
		return nil, fmt.Errorf("cannot create torrent engine : %s", err)
	}
	err = torrentEngine.InitTorrents()
	if err != nil {
		return nil, fmt.Errorf("cannot initilize torrents : %s", err)
	}

	return torrentEngine, nil
}
func InitDownloadEngine(db *db.Database) (*direct.Client, error) {
	defaultClientConfig, err := direct.NewClientDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("Cannot get default config : %s", err)
	}
	//initilize download client
	downloadClient, err := direct.CreateDownloadClient(defaultClientConfig, db)
	if err != nil {
		return nil, fmt.Errorf("Cannot torrent download client : %s", err)
	}
	err = downloadClient.InitDownloads()
	if err != nil {
		return nil, fmt.Errorf("Cannot initilize downloads : %s", err)
	}

	return downloadClient, nil
}
func (api *API) ExportOpenApi() {
	//write api json to file
	apiJson, err := json.Marshal(api.humaApi.OpenAPI())
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
}
func AddSystemRoutes(handler handlers.SystemHandler, humaApi huma.API) {
	// huma.Register(humaApi, huma.Operation{
	// 	OperationID: "get-system-info",
	// 	Method:      http.MethodGet,
	// 	Path:        "/system/info",
	// 	Summary:     "Get system info",
	// }, handler.GetSystemInfo)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-file-system-nodes",
		Method:      http.MethodPost,
		Path:        "/os/filesystem",
		Summary:     "Get file system nodes",
	}, handler.GetFileSystemNodes)
}
func AddTorrentRoutes(handler handlers.TorrentHandler, humaApi huma.API) {
	//register api routes
	// registering the download torrent route manually because it's a multipart/form-data request
	schema := humaApi.OpenAPI().Components.Schemas.Schema(reflect.TypeOf(handlers.DownloadTorrentReqBody{}), true, "DownloadTorrentReqBodyStruct")
	huma.Register(humaApi, huma.Operation{
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
	}, handler.DownloadTorrent)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-all-torrents",
		Method:      http.MethodGet,
		Path:        "/torrent",
		Summary:     "Get all torrents",
	}, handler.GetTorrents)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-torrent",
		Method:      http.MethodGet,
		Path:        "/torrent/{infohash}",
		Summary:     "Get torrent",
	}, handler.GetTorrent)
	huma.Register(humaApi, huma.Operation{
		OperationID: "pause-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent/pause",
		Summary:     "Pause torrent",
	}, handler.PauseTorrent)
	huma.Register(humaApi, huma.Operation{
		OperationID: "resume-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent/resume",
		Summary:     "Resume torrent",
	}, handler.ResumeTorrent)
	huma.Register(humaApi, huma.Operation{
		OperationID: "remove-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent/remove",
		Summary:     "Remove torrent",
	}, handler.RemoveTorrent)
	huma.Register(humaApi, huma.Operation{
		OperationID: "delete-torrent",
		Method:      http.MethodPost,
		Path:        "/torrent/delete",
		Summary:     "Delete torrent",
	}, handler.DeleteTorrent)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-torrent-meta-info-with-magnet",
		Method:      http.MethodPost,
		Path:        "/meta/magnet",
		Summary:     "Get torrent meta info with magnet",
	}, handler.GetMetaWithMagnet)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-torrent-meta-info-with-file",
		Method:      http.MethodPost,
		Path:        "/meta/file",
		Summary:     "Get torrent meta info with file",
	}, handler.GetMetaWithFile)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-torrents-total-speed",
		Method:      http.MethodGet,
		Path:        "/torrent/speed",
		Summary:     "Get torrents total speed",
	}, handler.GetTorrentsTotalSpeed)

}
func AddDownloadRoutes(handler handlers.DownloadHandler, humaApi huma.API) {
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-download-meta",
		Method:      http.MethodPost,
		Path:        "/download/meta",
		Summary:     "Get meta data of download",
	}, handler.GetDownloadMeta)
	huma.Register(humaApi, huma.Operation{
		OperationID: "download",
		Method:      http.MethodPost,
		Path:        "/download",
		Summary:     "Download with url",
	}, handler.Download)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-downloads",
		Method:      http.MethodGet,
		Path:        "/download",
		Summary:     "Get all downloads",
	}, handler.GetDownloads)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-download",
		Method:      http.MethodGet,
		Path:        "/download/{id}",
		Summary:     "Get download",
	}, handler.GetDownload)
	huma.Register(humaApi, huma.Operation{
		OperationID: "pause-download",
		Method:      http.MethodPost,
		Path:        "/download/pause",
		Summary:     "Pause download",
	}, handler.PauseDownload)
	huma.Register(humaApi, huma.Operation{
		OperationID: "resume-download",
		Method:      http.MethodPost,
		Path:        "/download/resume",
		Summary:     "Resume download",
	}, handler.ResumeDownload)
	huma.Register(humaApi, huma.Operation{
		OperationID: "remove-download",
		Method:      http.MethodPost,
		Path:        "/download/remove",
		Summary:     "Remove download",
	}, handler.RemoveDownload)
	huma.Register(humaApi, huma.Operation{
		OperationID: "delete-download-with-files",
		Method:      http.MethodPost,
		Path:        "/download/delete",
		Summary:     "Delete download with files",
	}, handler.DeleteDownload)
	huma.Register(humaApi, huma.Operation{
		OperationID: "get-downloads-total-speed",
		Method:      http.MethodGet,
		Path:        "/download/speed",
		Summary:     "Get downloads total speed",
	}, handler.GetDownloadsTotalSpeed)
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
