package protocol

type Reader interface {
	ReadMessage() error
}

type ReaderCallbacks struct {
	File   FileHandler
	Folder FolderHandler
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
