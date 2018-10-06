package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// block on channel
	select {
	case err := <-possibleErrors:
		for _, s := range servers {
			// propose PR with context timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.Shutdown(ctx)
		}
		log.Fatal(err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, http.StatusText(http.StatusOK))
}
