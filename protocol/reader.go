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

func (r *reader) ReadAll() error {
	var err error // read auth msg
	for err == nil {
		err = r.readNext()
	}
	return err
}

func (r *reader) readNext() error {
	msg, err := r.readMsg(&File{}, &Folder{})
	if err != nil {
		return fmt.Errorf("Failed to read message: %v", err)
	}

	if file, ok := msg.(*File); ok {
		return r.readFile(*file)
	}
	if folder, ok := msg.(*Folder); ok {
		go r.FolderHandler(*folder)
		return nil
	}
	return fmt.Errorf("Unexpected message type")
}

func (r *reader) readFile(file File) error {
	c := make(chan []byte) // buffer by num blocks
	file.Content = c
	go r.FileHandler(file)
	err := r.readContent(file.Size, c)
	if err != nil {
		return fmt.Errorf("Failed to read content for file %v %v: %v", file.ID, file.Name, err)
	}
	// read checksum
	return nil
}

func (r *reader) readMsg(msgTypes ...interface{}) (interface{}, error) {
	bytes, err := r.readMsgBytes()
	if err != nil {
		return nil, fmt.Errorf("Failed to read message: %v", err)
	}
	for _, msg := range msgTypes {
		err = json.Unmarshal(bytes, msg)
		if err == nil {
			return msg, nil
		}
	}
	return nil, fmt.Errorf("Failed to parse message: %v", bytes)
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
		return nil, fmt.Errorf("Failed to read message content: %v", err)
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
