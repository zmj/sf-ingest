package upload

import "context"

func NewUploader(host, authID string) Uploader {
	return &uploader{}
}

type uploader struct {
}

func (u *uploader) CreateFile(ctx context.Context, parentSfID, name string, content Content) (string, error) {
	return "fi123", nil
}

func (u *uploader) CreateFolder(ctx context.Context, parentSfID, name string) (string, error) {
	return "fo123", nil
}
