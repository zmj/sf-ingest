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
	var rcvErr error
	go func() {
		rcvErr = s.rcv.ReadAll()
		cancel()
	}()
	for rcvErr == nil {
		select {
		case sfAuth, ok := <-s.rcv.SfAuth():
			if !ok {
				break
			}
			u, err := upload.NewUploader(sfAuth.Host, sfAuth.AuthID)
			if err != nil {
				s.serverError(err)
				break
			}
			s.uploader = u
		case file, ok := <-s.rcv.Files():
			if !ok {
				break
			}
			s.startFileUpload(ctx, file)
		case folder, ok := <-s.rcv.Folders():
			if !ok {
				break
			}
			s.startFolderUpload(ctx, folder)
		case uploadRes := <-s.uploadResults:
			if uploadRes.err != nil {
				s.uploadError(uploadRes.ID, uploadRes.err)
				break
			}
			s.uploadSuccess(uploadRes.ID, uploadRes.sfID)
		case <-ctx.Done():
		}
	}
	fmt.Printf("session end: %v\n", rcvErr)
	s.serverError(rcvErr)
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
		content := upload.Content{Size: file.Size, Bytes: file.Content}
		sfID, err := s.uploader.CreateFile(ctx, parentSfID.getValue(), file.Name, content)
		s.uploadResults <- uploadResult{file.ID, sfID, err}
	}()
}

func (s *session) startFolderUpload(ctx context.Context, folder protocol.Folder) {
	folderSfID := s.addLazySfID(folder.ID)
	if folder.SfID != "" {
		folderSfID.setValue(folder.SfID)
		return
	}
	parentSfID := s.addLazySfID(folder.ParentID)
	go func() {
		sfID, err := s.uploader.CreateFolder(ctx, parentSfID.getValue(), folder.Name)
		if err == nil {
			folderSfID.setValue(sfID)
		}
		s.uploadResults <- uploadResult{folder.ID, sfID, err}
	}()
}

type uploadResult struct {
	ID   uint
	sfID string
	err  error
}

func (s *session) uploadSuccess(id uint, sfID string) {
	s.send.ItemDone(protocol.ItemDone{id, sfID})
}

func (s *session) uploadError(id uint, err error) {
	err = fmt.Errorf("Upload failed %v: %v", id, err)
	fmt.Printf("upload err: %v\n", err)
	s.serverError(err)
}

func (s *session) serverError(err error) {
	s.send.ServerError(protocol.ServerError{err.Error()})
}
