package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	SRV_PORT = "7727"

	//go:embed all:assets
	assetsDir   embed.FS
	assetsFS, _ = fs.Sub(assetsDir, "assets")

	//go:embed all:html
	htmlDir   embed.FS
	htmlFS, _ = fs.Sub(htmlDir, "html")
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
		Addr:         fmt.Sprintf(":%s", SRV_PORT),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHttpHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		log.Printf("WebApp now starting at http://127.0.0.1:%s\n", SRV_PORT)
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
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServerFS(assetsFS)))
	// mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsFilepath))))

	// views endpoints
	otelHandleFunc("/", RootHandler)

	// component functions as "controllers" endpoints
	otelHandleFunc("POST /controllers/calculate", CalculatePartialHandler)

	handler := otelhttp.NewHandler(mux, "/")
	return handler
}
