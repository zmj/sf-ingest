package protocol

type Reader interface {
	ReadAll() error
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
	ID       uint
	ParentID uint
	Name     string
}

type Writer interface {
	ItemDone(ItemDone) error
	Error(Error) error
}

type ItemDone struct {
	ID   uint
	SfID string
}

type Error struct {
	Message string
}
