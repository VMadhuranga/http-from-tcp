package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"log"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const crlf = "\r\n"

type Writer struct {
	Res io.Writer
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := "HTTP/1.1 "
	switch statusCode {
	case StatusOK:
		statusLine += fmt.Sprintf("%v OK", statusCode)
	case StatusBadRequest:
		statusLine += fmt.Sprintf("%v Bad Request", statusCode)
	case StatusInternalServerError:
		statusLine += fmt.Sprintf("%v Internal Server Error", statusCode)
	default:
		statusLine += fmt.Sprintf("%v ", statusCode)
	}

	_, err := w.Res.Write([]byte(statusLine + crlf))
	if err != nil {
		log.Printf("error writing status line: %v\n", err)
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defHeaders := headers.NewHeaders()
	defHeaders["Content-Length"] = strconv.Itoa(contentLen)
	defHeaders["Connection"] = "closed"
	defHeaders["Content-Type"] = "text/plain"

	return defHeaders
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for k, v := range headers {
		_, err := w.Res.Write([]byte(fmt.Sprintf("%v: %v%v", k, v, crlf)))
		if err != nil {
			log.Printf("error writing header: %v\n", err)
			return err
		}
	}

	_, err := w.Res.Write([]byte(crlf))
	if err != nil {
		log.Printf("error writing crlf: %v\n", err)
		return err
	}

	return nil
}

func (w *Writer) WriteBody(body []byte) error {
	_, err := w.Res.Write(body)
	if err != nil {
		log.Printf("error writing body: %v\n", err)
		return err
	}

	return nil
}
