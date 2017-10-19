package server

import (
	"fmt"
	"net"
)

type Server struct {
	port     int
	listener net.Listener
}

func (srv *Server) ListenAndServe() error {
	err := srv.startListener()
	if err != nil {
		return fmt.Errorf("Unable to start: %v", err)
	}
	defer srv.listener.Close()
	for err == nil {
		err = srv.acceptAndServe()
	}
	return err
}

func (srv *Server) startListener() error {
	if srv.listener != nil {
		return fmt.Errorf("Already listening")
	}
	if srv.port == 0 {
		srv.port = 9284
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", srv.port))
	if err != nil {
		return fmt.Errorf("Failed to start listener: %v", err)
	}
	srv.listener = listener
	return nil
}

func (srv *Server) acceptAndServe() error {
	conn, err := srv.listener.Accept()
	if err != nil {
		return fmt.Errorf("Connection failed: %v", err)
	}
	session := newSession(conn)
	go session.serve()
	return nil
}
