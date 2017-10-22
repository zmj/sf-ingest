package upload

import (
	"context"

	"github.com/zmj/sf-ingest/buffer"
)

type Uploader interface {
	CreateFile(ctx context.Context, parentSfID, name string, content Content) (string, error)
	CreateFolder(ctx context.Context, parentSfID, name string) (string, error)
}

type Content struct {
	Size  uint64
	Bytes <-chan *buffer.Buffer
	// checksum
	current *buffer.Buffer
}
