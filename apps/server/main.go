package main

import (
	"downite/api"
)

func main() {
	humaApi := api.ApiInit(&api.ApiOptions{
		Port: 9999,
	})

	humaApi.Run()
}
