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
	"fmt"
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

	handler.Routes["/"] = homeAction
	handler.Routes["/save"] = saveAction

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

//noinspection GoUnusedParameter
func homeAction(response http.ResponseWriter, request *http.Request) {
	io.WriteString(response,
		`<html>
		<body>
			<p>This is simple http server written in Go.</p>
			<form action="save" method="POST">
			    <label>Enter Text:</label>
				<input type="text" name="name">
				<input type="submit" value="Submit">
			</form>
		</body>
		</html>`)
}

//noinspection GoUnusedParameter
func saveAction(response http.ResponseWriter, request *http.Request) {
	template := `<html>
		<body>
			<p>This is simple http server written in Go.</p>
			<p>Parameter value: <b>%s</b></p>
		</body>
		</html>`
	request.ParseForm()

	name := request.PostForm.Get("name")

	body := fmt.Sprintf(template, name)
	io.WriteString(response, body)
}

//ServeHTTP routes incoming requests
func (handler *httpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	path := request.URL.Path

	if controller, ok := handler.Routes[path]; ok {
		controller(response, request)
	} else {
		response.WriteHeader(http.StatusNotFound)
		io.WriteString(response, "Page was not found: "+path)
		log.Printf("Not Found: '%s'", path)
	}
}
