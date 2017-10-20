package server

import (
	"context"
	"fmt"
	"net"

	"github.com/zmj/sf-ingest/protocol"
	"github.com/zmj/sf-ingest/upload"
)

func newSession(conn net.Conn) *session {
	s := &session{
		rcv:           protocol.NewReader(conn),
		send:          protocol.NewSender(conn),
		sfIDs:         make(map[uint]*lazySfID),
		uploadResults: make(chan uploadResult),
	}
	return s
}

type session struct {
	rcv           protocol.Receiver
	send          protocol.Sender
	sfIDs         map[uint]*lazySfID
	uploader      upload.Uploader
	uploadResults chan uploadResult
}

func (s *session) serve(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		err := s.rcv.ReadAll()
		fmt.Printf("session end: %v\n", err)
		cancel()
	}()
	for done := false; !done; {
		select {
		case sfAuth := <-s.rcv.SfAuth():
			s.uploader = upload.NewUploader(sfAuth.Host, sfAuth.AuthID)
		case file := <-s.rcv.Files():
			s.startFileUpload(ctx, file)
		case folder := <-s.rcv.Folders():
			s.startFolderUpload(ctx, folder)
		case uploadRes := <-s.uploadResults:
			if uploadRes.err != nil {
				s.uploadError(uploadRes.item, uploadRes.err)
			} else {
				s.uploadSuccess(uploadRes.item, uploadRes.sfID)
			}
		case <-ctx.Done():
			done = true
		}
	}
	for range s.uploadResults {
	}
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
	if lazy.sfID != "" {
		return
	}
	lazy.sfID = sfID
	close(lazy.done)
}

func (s *session) addLazySfID(id uint) *lazySfID {
	var sfID *lazySfID
	if sfID, exists := s.sfIDs[id]; exists {
		return sfID
	}
	sfID = &lazySfID{"", make(chan struct{})}
	s.sfIDs[id] = sfID
	return sfID
}

func (s *session) startFileUpload(ctx context.Context, file protocol.File) {
	parentSfID := s.addLazySfID(file.ParentID)
	go func() {
		// err for uploader nil
		sfID, err := s.uploader.CreateFile(ctx, parentSfID.getValue(), file.Name, upload.Content{file.Size, file.Content})
		s.uploadResults <- uploadResult{file.Item, sfID, err}
	}()
}

func (s *session) startFolderUpload(ctx context.Context, folder protocol.Folder) {
	parentSfID := s.addLazySfID(folder.ParentID)
	folderSfID := s.addLazySfID(folder.ID)
	go func() {
		// err for uploader nil
		sfID, err := s.uploader.CreateFolder(ctx, parentSfID.getValue(), folder.Name)
		if err == nil {
			folderSfID.setValue(sfID)
		}
		s.uploadResults <- uploadResult{folder.Item, sfID, err}
	}()
}

type uploadResult struct {
	item protocol.Item
	sfID string
	err  error
}

func (s *session) uploadSuccess(item protocol.Item, sfID string) {
	s.send.ItemDone(protocol.ItemDone{item.ID, sfID})
}

func (s *session) uploadError(item protocol.Item, err error) {
	msg := fmt.Sprintf("Upload failed %v (%v): %v", item.Name, item.ID, err)
	s.send.ServerError(protocol.ServerError{msg})
}
