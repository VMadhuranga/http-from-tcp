package request

import (
	"errors"
	"fmt"
	"io"
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
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	state       requestState
}

func parseRequestLine(rawRequest string) (RequestLine, int, error) {
	rawRequestSlice := strings.Split(rawRequest, crlf)
	if len(rawRequestSlice) <= 1 {
		return RequestLine{}, 0, nil
	}

	rawRequestLineSlice := strings.Split(rawRequestSlice[0], " ")

	method := rawRequestLineSlice[0]
	if method != strings.ToUpper(method) {
		return RequestLine{}, len(method), errors.New("invalid method")
	}

	requestTarget := rawRequestLineSlice[1]
	if requestTarget == "" {
		return RequestLine{},
			len(fmt.Sprintf("%v %v", method, requestTarget)),
			errors.New("invalid request target")
	}

	httpVersion := strings.Split(rawRequestLineSlice[2], "/")[1]
	if httpVersion != "1.1" {
		return RequestLine{},
			len(fmt.Sprintf("%v %v %v", method, requestTarget, rawRequestLineSlice[2])),
			errors.New("invalid http version")
	}

	return RequestLine{
			Method:        method,
			RequestTarget: requestTarget,
			HttpVersion:   httpVersion,
		},
		len(rawRequestSlice[0]),
		nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		reqLine, n, err := parseRequestLine(string(data))
		if err != nil {
			return n, err
		}

		if n == 0 {
			return 0, nil
		}

		r.RequestLine = reqLine
		r.state = requestStateDone

		return n, nil
	case requestStateDone:
		return 0, errors.New("trying to read data in a done state")
	}

	return 0, errors.New("unknown state")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := &Request{
		state: requestStateInitialized,
	}

	for req.state != requestStateDone {
		if len(buf) == cap(buf) {
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
			return nil, err
		}

		readToIndex += n

		n, err = req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		newBuf := make([]byte, len(buf))
		copy(newBuf, buf[:readToIndex])
		buf = newBuf

		readToIndex -= n
	}

	return req, nil
}
