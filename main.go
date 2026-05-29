package main

import (
	"log"
	"net/http"
)

func main() {
	engine, err := NewFileEngine("events.log")
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	api := NewAPI(engine)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	addr := ":8080"
	log.Printf("Server running on %s\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
