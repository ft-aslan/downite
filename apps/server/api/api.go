package api

import (
	"downite/handlers"
	"fmt"

	"github.com/go-fuego/fuego"
)

func ApiInit() error {
	fmt.Println("Starting Downite server...")

	s := fuego.NewServer()

	// TODO(fatih): add scalar documentation
	// fuego.Get(s, "/docs", func(c *fuego.ContextNoBody) (fuego.HTML, error) {
	// 	return c.Render("docgen/scalar.html", nil)
	// })

	apiRoutes := fuego.Group(s, "/api/v1")

	fuego.Get(apiRoutes, "/torrent", handlers.GetTorrents)
	fuego.Post(apiRoutes, "/torrent-meta", handlers.GetTorrentMeta)

	err := s.Run()
	return err
}
