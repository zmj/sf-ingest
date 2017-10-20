package upload

import "context"

type Uploader interface {
	CreateFile(ctx context.Context, parentSfID, name string, content Content) (string, error)
	CreateFolder(ctx context.Context, parentSfID, name string) (string, error)
}

type Content struct {
	Size  uint64
	Bytes <-chan []byte
	// checksum
}
