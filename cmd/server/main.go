package main

import (
	"log"
	"net/http"
	"os"

	"buoy-calibration/internal/api"
	"buoy-calibration/internal/service"
	"buoy-calibration/internal/store"
)

func main() {
	addr := os.Getenv("BUOY_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	svc := service.New(store.NewMemory())
	server := api.NewServer(svc)
	log.Printf("buoy calibration service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.Routes()))
}
