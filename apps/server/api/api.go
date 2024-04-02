package api

import (
	"downite/handlers"
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/rs/cors"
)

func ApiInit() error {
	fmt.Println("Starting Downite server...")

	s := fuego.NewServer(fuego.WithCorsMiddleware(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}).Handler))

	// TODO(fatih): add scalar documentation
	// fuego.Get(s, "/docs", func(c *fuego.ContextNoBody) (fuego.HTML, error) {
	// 	return c.Render("docgen/scalar.html", nil)
	// })

	apiRoutes := fuego.Group(s, "/api/v1")

	fuego.Get(apiRoutes, "/torrent", handlers.GetTorrents)
	fuego.Post(apiRoutes, "/torrent", handlers.DownloadTorrent)
	fuego.Get(apiRoutes, "/torrent/:hash", handlers.GetTorrent)
	fuego.Post(apiRoutes, "/torrent-meta", handlers.GetTorrentMeta)

	err := s.Run()
	return err
}
