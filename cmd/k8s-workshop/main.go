package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kamenim/k8s-workshop/internal/diagnostics"
)

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

	go func() {
		server := &http.Server{
			Addr:    ":" + srvPort,
			Handler: router,
		}
		err := server.ListenAndServe()
		if err != nil {
			possibleErrors <- err
		}
	}()

	go func() {
		diagnostics := diagnostics.NewDiagnostics()
		diagServer := &http.Server{
			Addr:    ":" + diagPort,
			Handler: diagnostics,
		}
		err := diagServer.ListenAndServe()
		// err := http.ListenAndServe(":"+diagPort, diagnostics)
		if err != nil {
			// log.Fatal(err)
			possibleErrors <- err
		}
	}()

	// block on channel
	select {
	case err := <-possibleErrors:
		log.Fatal(err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, http.StatusText(http.StatusOK))
}
