package protocol

type Reader interface {
	ReadMessage() error
}

type ReaderCallbacks struct {
	FileHandler   FileHandler
	FolderHandler FolderHandler
}

type FileHandler func(File)
type FolderHandler func(Folder)

type File struct {
	Item
	Size    uint64
	Content <-chan []byte
}

type Folder struct {
	Item
}

type Item struct {
	ID       int
	ParentID int
	Name     string
}
