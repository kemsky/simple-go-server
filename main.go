package main

import (
	"net/http"
	"time"
	"log"
	"io"
	"os"
	"os/signal"
	"context"
	"syscall"
)

const readTimeout = 1000 * time.Second
const writeTimeout = 1000 * time.Second
const address = "0.0.0.0:8080"

type httpHandler struct {
	Routes map[string]func(http.ResponseWriter, *http.Request)
}

func main() {
	log.Printf("Starting server on '%s'...\n", address)

	handler := httpHandler{
		Routes: make(map[string]func(http.ResponseWriter, *http.Request)),
	}

	handler.Routes["/"] = homeController

	server := &http.Server{
		Addr:           address,
		Handler:        &handler,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	go start(server)

	wait()

	stop(server)
}

func start(server *http.Server) {
	log.Fatal(server.ListenAndServe())
}

func stop(server *http.Server) {
	log.Println("Stopping server...")
	time.Sleep(time.Millisecond * 500)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	log.Println("Server stopped")
}

func wait() {
	signals := make(chan os.Signal)

	signal.Notify(signals, syscall.SIGTERM)
	signal.Notify(signals, syscall.SIGINT)

	<-signals
	log.Println("Iterrupted...")
}

func homeController(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "This is simple http server written in Go.")
}

//ServeHTTP routes incoming requests
func (handler *httpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	url := request.URL.String()

	if controller, ok := handler.Routes[url]; ok {
		controller(response, request)
	} else {
		response.WriteHeader(http.StatusNotFound)
		io.WriteString(response, "Page was not found: "+url)
		log.Printf("Not Found: '%s'", url)
	}
}
