package protocol

type Receiver interface {
	ReadAll() error
	SfAuth() <-chan SfAuth
	Files() <-chan File
	Folders() <-chan Folder
}

type File struct {
	Item
	Size    uint64
	Content <-chan []byte
}

type Folder struct {
	Item
	SfID string
}

type Item struct {
	ID       uint
	ParentID uint
	Name     string
}

type SfAuth struct {
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
