package routes

import (
	"ilmavridis/url-shortener/config"
	"ilmavridis/url-shortener/middleware"

	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Configures router and returns server
func New() *http.Server {

	router := mux.NewRouter()

	router.HandleFunc("/", middleware.Logger(home)).Methods("GET")
	router.HandleFunc("/images/{imageName}", middleware.Logger(ReturnImage)).Methods("GET") // Returns images required from home handler for html page
	router.HandleFunc("/info/{shortUrl}", middleware.Logger(Info)).Methods("GET")
	router.HandleFunc("/short", middleware.Logger(ShortenUrl)).Methods("POST")
	router.HandleFunc("/{shortUrl}", middleware.Logger(ResolveUrl)).Methods("GET")
	router.NotFoundHandler = middleware.Logger(My404Handler)

	conf := config.Get()

	srv := &http.Server{
		Addr: conf.Server.Address,
		// Good practice to set timeouts to avoid Slowloris DDoS attacks.
		WriteTimeout: conf.Server.TimeoutWrite,
		ReadTimeout:  conf.Server.TimeoutRead,
		IdleTimeout:  conf.Server.TimeoutIdle,
		// Pass our instance of gorilla/mux in.
		Handler: router,
	}

	return srv
}

// Runs the server as a goroutine
func Run(srv *http.Server) <-chan error {
	errChannel := make(chan error)

	go func() {
		defer close(errChannel)
		if err := srv.ListenAndServe(); err != nil {
			errChannel <- err
		}
	}()

	return errChannel
}

func SetupGracefulShutdown(srv *http.Server) error {
	// We created a context with cancel() callback function. The cancel() function
	// is called once an OS interrupt is received.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	// The Shutdown() method is used to close the server gracefully by stopping receiving new requests
	// and completing the already processed requests. The Shutdown() allows current requests to be completed
	// in 15 * time.Second, set in shutdownCtx.
	err := srv.Shutdown(shutdownCtx)

	return err
}

func My404Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	// Returns response in json
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	jsonResp, _ := json.Marshal(map[string]string{"error": "404 page not found"})

	w.Write(jsonResp)
	return
}
