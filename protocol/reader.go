package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

func NewReader(conn io.Reader, cb ReaderCallbacks) Interface {
	return &reader{
		conn:          conn,
		fileHandler:   cb.File,
		folderHandler: cb.Folder,
		msgBuffer:     make([]byte, 4096),
	}
}

type reader struct {
	conn          io.Reader
	fileHandler   FileHandler
	folderHandler FolderHandler
	msgBuffer     []byte
}

func (r *reader) ReadMessage() error {
	msg, err := r.readMsg()
	if err != nil {
		break
	}

	var file File
	err = json.Unmarshal(msg, &file)
	if err == nil {
		c := make(chan []byte) // buffer by num blocks
		file.Content = c
		go r.fileHandler(file)
		return r.readContent(file.Size, c)
	}

	var folder Folder
	err = json.Unmarshal(msg, &folder)
	if err == nil {
		go r.folderHandler(folder)
		return nil
	}

	return fmt.Errorf("Unknown message type")
}

func (r *reader) readMsg() ([]byte, error) {
	size, err := r.readMsgSize()
	if err != nil {
		return nil, fmt.Errorf("Failed to read message size: %v", err)
	}
	bufLen := len(r.msgBuffer)
	if int(size) > bufLen {
		return nil, fmt.Errorf("Message size %v exceeds buffer size %v", size, bufLen)
	}
	buf := r.msgBuffer[:size]
	_, err = io.ReadFull(r.conn, buf)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message: %v", err)
	}
	return buf, nil
}

func (r *reader) readMsgSize() (uint16, error) {
	var size uint16
	err := binary.Read(r.conn, binary.BigEndian, &size)
	return size, err
}

func (r *reader) readContent(n uint64, c chan<- []byte) error {
	var read uint64
	for read < n {
		buf := make([]byte, 4096) // buffer pool
		_, err := io.ReadFull(r.conn, buf)
		if err != nil {
			return fmt.Errorf("Failed to read content block: %v", err)
		}
		c <- buf
		read += uint64(len(buf))
	}
	return nil
}
