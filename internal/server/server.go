package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type HandlerError struct {
	StatusCode    response.StatusCode
	StatusMessage string
}

func (hErr *HandlerError) write(w io.Writer) error {
	_, err := w.Write([]byte(fmt.Sprintf("%v %v", hErr.StatusCode, hErr.StatusMessage)))
	if err != nil {
		log.Printf("error writing handler error: %v\n", err)
		return err
	}

	return nil
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	Port      int
	listener  net.Listener
	connState atomic.Bool
	handler   Handler
}

func (s *Server) Close() error {
	s.connState.CompareAndSwap(true, false)
	return s.listener.Close()
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("error getting request from connection: %v\n", err)
		hErr := &HandlerError{
			StatusCode:    response.StatusInternalServerError,
			StatusMessage: err.Error(),
		}
		hErr.write(conn)
		return
	}

	resWriter := response.Writer{
		Res: conn,
	}
	s.handler(&resWriter, req)
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if !s.connState.Load() {
			return
		}
		if err != nil {
			log.Printf("error waiting for next connection: %v\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Printf("error announcing local network address: %v\n", err)
		return nil, err
	}

	s := &Server{
		Port:     port,
		listener: listener,
		handler:  handler,
	}
	s.connState.Store(true)

	go s.listen()

	return s, nil
}
