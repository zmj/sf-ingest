package upload

import (
	"context"
	"fmt"
)

func NewUploader(host, authID string) Uploader {
	return &uploader{}
}

type uploader struct {
}

func (u *uploader) CreateFile(ctx context.Context, parentSfID, name string, content Content) (string, error) {
	if u == nil {
		return "", fmt.Errorf("Uploader not initialized")
	}
	for b := range content.Bytes {
		fmt.Printf("Fake upload %v\n%v", name, string(b))
	}
	return "fi123", nil
}

func (u *uploader) CreateFolder(ctx context.Context, parentSfID, name string) (string, error) {
	if u == nil {
		return "", fmt.Errorf("Uploader not initialized")
	}
	return "fo123", nil
}
