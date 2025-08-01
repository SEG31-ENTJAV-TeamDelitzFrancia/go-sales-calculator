package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// handle SIGINT (Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// setup opentelemetry
	otelShut, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}

	// handle otelShut properly
	defer func() {
		err = errors.Join(err, otelShut(context.Background()))
	}()

	// setup server
	srv := &http.Server{
		Addr:         ":7727",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHttpHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// wait for interruption
	select {
	case err = <-srvErr:
		// error when starting HTTP Server
		return
	case <-ctx.Done():
		// wait for first Ctrl+C
		// stop receiving signal notifications as soon as possible
		stop()
	}

	// when shut is called, ListenAndServe immediately returned EndServerClosed
	err = srv.Shutdown(context.Background())
	return
}

func newHttpHandler() http.Handler {
	mux := http.NewServeMux()

	// replace mux.HandleFunc to otelHandleFunc
	otelHandleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// assets content (e.g. assets)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsFilepath))))

	// views endpoints
	otelHandleFunc("/", RootHandler)

	// component functions as "controllers" endpoints
	otelHandleFunc("POST /controllers/calculate", CalculatePartialHandler)

	handler := otelhttp.NewHandler(mux, "/")
	return handler
}
