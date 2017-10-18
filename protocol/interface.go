package protocol

import "io"

type Interface interface {
	Start(io.ReadWriteCloser) // buffer pool
	ItemDone(int)
	FileHandler(FileHandler)
	FolderHandler(FolderHandler)
}

type FileHandler func(File)
type FolderHandler func(Folder)

type File struct {
	Id       int
	ParentId int
	Name     string
	Size     int64
	Content  <-chan []byte
}

type Folder struct {
	Id       int
	ParentId int
	Name     string
}
