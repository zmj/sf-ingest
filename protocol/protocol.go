package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

func New() Interface {
	return &protocol{}
}

type protocol struct {
	conn          io.ReadWriter
	fileHandler   FileHandler
	folderHandler FolderHandler
	msgBuffer     []byte
}

func (p *protocol) read() {
	var err error
	for err != nil {
		msg, err := p.readMsg()
		if err != nil {
			break
		}

		var file File
		err = json.Unmarshal(msg, &file)
		if err == nil {
			c := make(chan []byte) // buffer by num blocks
			file.Content = c
			go p.fileHandler(file)
			err = p.readContent(file.Size, c)
			continue
		}

		var folder Folder
		err = json.Unmarshal(msg, &folder)
		if err == nil {
			go p.folderHandler(folder)
			continue
		}

		err = fmt.Errorf("Unknown message type")
	}
}

func (p *protocol) readMsg() ([]byte, error) {
	size, err := p.readMsgSize()
	if err != nil {
		return nil, fmt.Errorf("Failed to read message size: %v", err)
	}
	if int(size) > len(p.msgBuffer) {
		return nil, fmt.Errorf("Message size %v exceeds buffer size %v", size, len(p.msgBuffer))
	}
	buf := p.msgBuffer[:size]
	_, err = io.ReadFull(p.conn, buf)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message: %v", err)
	}
	return buf, nil
}

func (p *protocol) readMsgSize() (uint16, error) {
	var size uint16
	err := binary.Read(p.conn, binary.BigEndian, &size)
	return size, err
}

func (p *protocol) readContent(n uint64, c chan<- []byte) error {
	var read uint64
	for read < n {
		buf := make([]byte, 4096) // buffer pool
		_, err := io.ReadFull(p.conn, buf)
		if err != nil {
			return fmt.Errorf("Failed to read content block: %v", err)
		}
		c <- buf
		read += uint64(len(buf))
	}
	return nil
}

func (p *protocol) Start(conn io.ReadWriter) {
	p.conn = conn
	go p.read()
}

func (p *protocol) ItemDone(item ItemDone) {
}

func (p *protocol) FileHandler(fh FileHandler) {
	p.fileHandler = fh
}

func (p *protocol) FolderHandler(fh FolderHandler) {
	p.folderHandler = fh
}
