package protocol

import "io"

type Reader interface {
	Start(io.ReadWriter) // buffer pool
	ItemDone(ItemDone)
	FileHandler(FileHandler)
	FolderHandler(FolderHandler)
}

type FileHandler func(File)
type FolderHandler func(Folder)

type File struct {
	Id       int
	ParentId int
	Name     string
	Size     uint64
	Content  <-chan []byte
}

type Folder struct {
	Id       int
	ParentId int
	Name     string
}

type ItemDone struct {
	Id   int
	SfId string
}
