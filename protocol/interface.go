package protocol

import "github.com/zmj/sf-ingest/buffer"

type Receiver interface {
	ReadAll() error
	SfAuth() <-chan SfAuth
	Files() <-chan File
	Folders() <-chan Folder
}

type File struct {
	T        string
	ID       uint
	ParentID uint
	Name     string
	Size     uint64
	Content  <-chan *buffer.Buffer
}

type Folder struct {
	T        string
	ID       uint
	ParentID uint
	Name     string
	SfID     string
}

type SfAuth struct {
	T      string
	Host   string
	AuthID string
}

type Sender interface {
	ItemDone(ItemDone) error
	ServerError(ServerError) error
}

type ItemDone struct {
	ID   uint
	SfID string
}

type ServerError struct {
	Message string
}
