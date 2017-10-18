package server

import (
	"fmt"
	"net"

	"github.com/zmj/sf-ingest/protocol"
)

func NewSession(port int) (*Session, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, fmt.Errorf("Failed to start listener: %v", err)
	}
	conn, err := listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("Failed to open connection: %v", err)
	}
	return &Session{conn}, nil
}

type Session struct {
	conn net.Conn
}

func (s *Session) ListenOne() error {
	r := protocol.NewReader(s.conn, protocol.ReaderCallbacks{s.fileHandler, s.folderHandler})
	return r.ReadMessage()
}

func (s *Session) fileHandler(file protocol.File) {

}

func (s *Session) folderHandler(folder protocol.Folder) {

}
