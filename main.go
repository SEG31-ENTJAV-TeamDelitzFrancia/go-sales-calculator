package main

import (
	"log"
	"net/http"
)

var (
	filepath       = "./html"
	assetsFilepath = "./assets"
)

var (
	AdditionalSalesCommission  = 0.075
	SalesCodeAlphaRate         = 0.5
	SalesCodeAlphaAdditional   = 100.0
	SalesCodeBravoRate         = 0.2
	SalesCodeBravoAdditional   = 100.0
	SalesCodeCharlieRate       = 0.25
	SalesCodeCharlieAdditional = 50.0
	MinimumCommission          = 2500.0
)

func main() {
	srv := http.NewServeMux()

	// assets content (e.g. assets)
	srv.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsFilepath))))

	// views endpoints
	srv.HandleFunc("/", RootHandler)

	// component functions as "controllers" endpoints
	srv.HandleFunc("POST /controllers/calculate", CalculatePartialHandler)

	// serve
	log.Println("Web App now running at http://localhost:7727")
	if err := http.ListenAndServe(":7727", srv); err != nil {
		log.Fatalln("Something went wrong when running server.", err)
	}
}
