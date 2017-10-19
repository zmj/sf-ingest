package server

import (
	"fmt"
	"net"

	"github.com/zmj/sf-ingest/protocol"
)

func newSession(conn net.Conn) *session {
	s := &session{
		conn:        conn,
		newFile:     make(chan protocol.File),
		newFolder:   make(chan protocol.Folder),
		upload:      make(chan uploadResult),
		readErr:     make(chan error),
		parentSfIDs: make(map[uint]*lazySfID),
	}
	s.rdr = protocol.NewReader(conn, protocol.ReaderCallbacks{s.fileHandler, s.folderHandler})
	s.wr = protocol.NewWriter(conn)
	return s
}

type session struct {
	conn        net.Conn
	wr          protocol.Writer
	rdr         protocol.Reader
	newFile     chan protocol.File
	newFolder   chan protocol.Folder
	upload      chan uploadResult
	readErr     chan error
	err         error
	parentSfIDs map[uint]*lazySfID
}

func (s *session) serve() {
	defer s.conn.Close()
	for s.err == nil {
		select {
		case file := <-s.newFile:
			s.addParentID(file.ParentID)
			// start upload
		case folder := <-s.newFolder:
			s.addParentID(folder.ParentID)
			s.addParentID(folder.ID)
			// start upload
		case upload := <-s.upload:
			if upload.err != nil {
				s.err = upload.err
				break
			}
			if lazyID, exists := s.parentSfIDs[upload.item.ID]; exists {
				lazyID.setValue(upload.sfID)
			}
			s.err = s.wr.ItemDone(protocol.ItemDone{upload.item.ID, upload.sfID})
		case readErr := <-s.readErr:
			s.err = readErr
		}
	}
	s.wr.Error(protocol.Error{s.err.Error()})
	fmt.Printf("session end: %v\n", s.err)
}

func (s *session) addParentID(id uint) {
	if _, exists := s.parentSfIDs[id]; exists {
		return
	}
	s.parentSfIDs[id] = &lazySfID{"", make(chan struct{})}
}

func (s *session) fileHandler(file protocol.File) {
	if s.err != nil {
		return
	}
	s.newFile <- file
}

func (s *session) folderHandler(folder protocol.Folder) {
	if s.err != nil {
		return
	}
	s.newFolder <- folder
}

type uploadResult struct {
	item protocol.Item
	sfID string
	err  error
}

type lazySfID struct {
	sfID string
	done chan struct{}
}

func (lazy *lazySfID) getValue() string {
	<-lazy.done
	return lazy.sfID
}

func (lazy *lazySfID) setValue(sfID string) {
	lazy.sfID = sfID
	close(lazy.done)
}
