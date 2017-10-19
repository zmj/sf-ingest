package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

func NewReader(rdr io.Reader, cb ReaderCallbacks) Reader {
	return &reader{
		ReaderCallbacks: cb,
		rdr:             rdr,
		msgBuffer:       make([]byte, 4096),
	}
}

type reader struct {
	ReaderCallbacks
	rdr       io.Reader
	msgBuffer []byte
}

func (r *reader) ReadMessage() error {
	msg, err := r.readMsgBytes()
	if err != nil {
		return fmt.Errorf("Failed to read message: %v", err)
	}

	var file File
	err = json.Unmarshal(msg, &file)
	if err == nil {
		c := make(chan []byte) // buffer by num blocks
		file.Content = c
		go r.FileHandler(file)
		return r.readContent(file.Size, c)
	}

	var folder Folder
	err = json.Unmarshal(msg, &folder)
	if err == nil {
		go r.FolderHandler(folder)
		return nil
	}

	return fmt.Errorf("Unknown message type")
}

func (r *reader) readMsgBytes() ([]byte, error) {
	size, err := r.readMsgSize()
	if err != nil {
		return nil, fmt.Errorf("Failed to read message size: %v", err)
	}
	bufLen := len(r.msgBuffer)
	if int(size) > bufLen {
		return nil, fmt.Errorf("Message size %v exceeds buffer size %v", size, bufLen)
	}
	buf := r.msgBuffer[:size]
	_, err = io.ReadFull(r.rdr, buf)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message: %v", err)
	}
	return buf, nil
}

func (r *reader) readMsgSize() (uint16, error) {
	var size uint16
	err := binary.Read(r.rdr, binary.BigEndian, &size)
	return size, err
}

func (r *reader) readContent(n uint64, c chan<- []byte) error {
	defer close(c)
	if n == 0 {
		return nil
	}
	var read uint64
	for read < n {
		var toRead uint64 = 4096
		if n-read < toRead {
			toRead = n - read
		}
		buf := make([]byte, toRead) // buffer pool

		_, err := io.ReadFull(r.rdr, buf)
		if err != nil {
			return fmt.Errorf("Failed to read content block: %v", err)
		}
		c <- buf
		read += uint64(len(buf))
	}
	return nil
}
