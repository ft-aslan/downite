package api

import (
	"downite/handlers"
	"fmt"

	"github.com/go-fuego/fuego"
)

func ApiInit() error {
	fmt.Println("Starting Downite server...")

	s := fuego.NewServer()

	apiRoutes := fuego.Group(s, "/api/v1")

	fuego.Get(apiRoutes, "/torrents", handlers.GetTorrents)

	err := s.Run()
	return err
}
