package main

import (
	"go-chat/endpoint"
	"go-chat/router"
)

func main() {
	go endpoint.RoomSet.Run()
	r := router.InitRouter()
	if err := r.Run("0.0.0.0:8000"); err != nil {
		panic(err)
	}
}
