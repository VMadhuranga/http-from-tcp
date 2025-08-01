package request

import (
	"errors"
	"httpfromtcp/internal/headers"
	"io"
	"log"
	"strconv"
	"strings"
)

const bufferSize = 8
const crlf = "\r\n"

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       requestState
}

func parseRequestLine(rawRequestLine string) (RequestLine, int, error) {
	crlfIDX := strings.Index(rawRequestLine, crlf)
	if crlfIDX == -1 {
		return RequestLine{}, 0, nil
	}

	requestLine := rawRequestLine[:crlfIDX]
	requestLineSlice := strings.Split(requestLine, " ")

	method := requestLineSlice[0]
	if method != strings.ToUpper(method) {
		return RequestLine{}, 0, errors.New("invalid method")
	}

	requestTarget := requestLineSlice[1]
	if requestTarget == "" {
		return RequestLine{}, 0, errors.New("invalid request target")
	}

	httpVersion := strings.Split(requestLineSlice[2], "/")[1]
	if httpVersion != "1.1" {
		return RequestLine{}, 0, errors.New("invalid http version")
	}

	return RequestLine{
			Method:        method,
			RequestTarget: requestTarget,
			HttpVersion:   httpVersion,
		},
		len(requestLine + crlf),
		nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		reqLine, n, err := parseRequestLine(string(data))
		if err != nil {
			log.Printf("error parsing request line: %v\n", err)
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = reqLine
		r.state = requestStateParsingHeaders

		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(string(data))
		if err != nil {
			log.Printf("error parsing headers: %v\n", err)
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}

		return n, nil
	case requestStateParsingBody:
		contentLengthVal := r.Headers.Get("content-length")
		if contentLengthVal == "" {
			r.state = requestStateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(contentLengthVal)
		if err != nil {
			log.Printf("error converting string to int: %v\n", err)
			return 0, err
		}
		if contentLength < 0 {
			return 0, errors.New("invalid content length value")
		}

		r.Body = append(r.Body, data...)
		if len(r.Body) > contentLength {
			return 0, errors.New("body length is greater than content length")
		}
		if len(r.Body) == contentLength {
			r.state = requestStateDone
		}

		return len(data), nil
	case requestStateDone:
		return 0, errors.New("trying to read data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			log.Printf("error parsing single: %v\n", err)
			return 0, err
		}
		if n == 0 {
			break
		}

		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := &Request{
		Headers: headers.NewHeaders(),
		state:   requestStateInitialized,
	}

	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			req.state = requestStateDone
			break
		}
		if err != nil {
			log.Printf("error reading request: %v\n", err)
			return nil, err
		}

		readToIndex += n

		n, err = req.parse(buf[:readToIndex])
		if err != nil {
			log.Printf("error parsing request: %v\n", err)
			return nil, err
		}

		copy(buf, buf[n:readToIndex])

		readToIndex -= n
	}

	contentLengthVal := req.Headers.Get("content-length")
	if req.state == requestStateDone && contentLengthVal != "" {
		contentLength, err := strconv.Atoi(req.Headers.Get("content-length"))
		if err != nil {
			log.Printf("error parsing string to int: %v\n", err)
			return nil, err
		}
		if len(req.Body) < contentLength {
			return nil, errors.New("body length is less than content length")
		}
	}

	return req, nil
}
