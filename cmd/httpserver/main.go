package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069

func httpBinProxyHandler(w *response.Writer, r *request.Request) {
	path := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin")
	url := "https://httpbin.org" + path

	res, err := http.Get(url)
	if err != nil {
		log.Printf("error getting response: %v", err)
		return
	}
	defer res.Body.Close()

	defHeaders := response.GetDefaultHeaders(0)
	delete(defHeaders, "Content-Length")
	defHeaders["Transfer-Encoding"] = "chunked"
	defHeaders["Trailer"] = "X-Content-SHA256, X-Content-Length"

	w.WriteStatusLine(response.StatusCode(res.StatusCode))
	w.WriteHeaders(defHeaders)

	chunk := make([]byte, 1024)
	rawBody := []byte{}
	for {
		n, err := res.Body.Read(chunk)
		if err == io.EOF {
			w.WriteChunkedBodyDone()
			break
		}
		if err != nil {
			log.Printf("error reading chunk: %v", err)
			return
		}
		fmt.Println("data read:", n)

		rawBody = append(rawBody, chunk[:n]...)
		w.WriteChunkedBody(chunk[:n])
	}

	trailers := headers.NewHeaders()

	sum := sha256.Sum256(rawBody)
	trailers["X-Content-SHA256"] = string(sum[:])
	trailers["X-Content-Length"] = strconv.Itoa(len(rawBody))

	w.WriteTrailers(trailers)
}

var resBody200 = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
var resBody400 = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
var resBody500 = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

func handler(w *response.Writer, r *request.Request) {
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin") {
		httpBinProxyHandler(w, r)
		return
	}

	resStatus := response.StatusOK
	resBody := resBody200
	defHeaders := response.GetDefaultHeaders(len(resBody))

	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		resStatus = response.StatusBadRequest
		resBody = resBody400
		defHeaders["Content-Length"] = strconv.Itoa(len(resBody))
	case "/myproblem":
		resStatus = response.StatusInternalServerError
		resBody = resBody500
		defHeaders["Content-Length"] = strconv.Itoa(len(resBody))
	case "/video":
		data, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Printf("error reading file: %v", err)
			resBody = resBody400
		} else {
			defHeaders["Content-Type"] = "video/mp4"
			resBody = string(data)
		}
		defHeaders["Content-Length"] = strconv.Itoa(len(resBody))
	}

	w.WriteStatusLine(resStatus)
	w.WriteHeaders(defHeaders)
	w.WriteBody([]byte(resBody))
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	server, err := server.Serve(port, handler)
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
