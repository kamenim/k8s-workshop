package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kamenim/k8s-workshop/internal/diagnostics"
)

type serverConf struct {
	port   string
	router http.Handler
	name   string
}

func main() {
	log.Print("Simple hello")

	srvPort := os.Getenv("PORT")
	if len(srvPort) == 0 {
		srvPort = "8080"
	}
	diagPort := os.Getenv("DIAG_PORT")
	if len(diagPort) == 0 {
		diagPort = "8585"
	}

	router := mux.NewRouter()
	router.HandleFunc("/", hello)

	possibleErrors := make(chan error, 2)

	configurations := []serverConf{
		{
			port:   srvPort,
			router: router,
			name:   "application server",
		},
		{
			port:   diagPort,
			router: diagnostics.NewDiagnostics(),
			name:   "Diag server",
		},
	}

	servers := make([]*http.Server, 2)

	for i, c := range configurations {
		go func(conf serverConf, i int) {
			log.Printf("Starting server %s", conf.name)
			servers[i] = &http.Server{
				Addr:    ":" + conf.port,
				Handler: conf.router,
			}
			err := servers[i].ListenAndServe()
			if err != nil {
				possibleErrors <- err
			}
		}(c, i)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-possibleErrors:
		log.Printf("Got an error: %v", err)
	case sig := <-interrupt:
		log.Printf("Recevied the signal %v", sig)
	}

	for _, s := range servers {
		timeout := 5 * time.Second
		log.Printf("Shutdown with timeout: %s", timeout)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := s.Shutdown(ctx)
		if err != nil {
			fmt.Println(err)
		}
		log.Printf("Server gracefully stopped")
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Print("The hello handler was called")
	fmt.Fprint(w, http.StatusText(http.StatusOK))
}
