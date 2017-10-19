package server

import (
	"fmt"
	"net"

	"github.com/zmj/sf-ingest/protocol"
)

type session struct {
	conn net.Conn
}

func (s *session) serve() {
	rdr := protocol.NewReader(s.conn, protocol.ReaderCallbacks{s.fileHandler, s.folderHandler})
	err := rdr.ReadAll()
	fmt.Printf("err: %v\n", err)
}

func (s *session) fileHandler(file protocol.File) {
	fmt.Printf("file %v\n", file.Name)
	for b := range file.Content {
		fmt.Printf("content %v\n", b)
	}
}

func (s *session) folderHandler(folder protocol.Folder) {
	fmt.Println("folder")
}
