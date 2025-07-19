package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func testHandler(w *response.Writer, req *request.Request) {
	resStatus := response.StatusOK
	resBody := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		resStatus = response.StatusBadRequest
		resBody = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

	case "/myproblem":
		resStatus = response.StatusInternalServerError
		resBody = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
	}

	w.WriteStatusLine(resStatus)

	defHeaders := response.GetDefaultHeaders(len(resBody))
	w.WriteHeaders(defHeaders)

	w.WriteBody([]byte(resBody))
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	server, err := server.Serve(port, testHandler)
	if err != nil {
		log.Fatalf("error starting server: %v\n", err)
	}
	defer server.Close()
	log.Println("server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("server gracefully stopped")
}
