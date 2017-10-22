package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

func NewReader(rdr io.Reader) Receiver {
	return &receiver{
		rdr:       rdr,
		sfAuth:    make(chan SfAuth),
		files:     make(chan File),
		folders:   make(chan Folder),
		msgBuffer: make([]byte, 4096),
	}
}

type receiver struct {
	rdr       io.Reader
	msgBuffer []byte
	sfAuth    chan SfAuth
	files     chan File
	folders   chan Folder
}

func (r *receiver) ReadAll() error {
	defer close(r.sfAuth)
	defer close(r.files)
	defer close(r.folders)
	// require auth msg first?
	var err error
	for err == nil {
		err = r.readNext()
	}
	return err
}

func (r *receiver) readNext() error {
	msg, err := r.readMsg(&File{}, &Folder{}, &SfAuth{})
	if err != nil {
		return fmt.Errorf("Failed to read message: %v", err)
	}

	if file, ok := msg.(*File); ok {
		return r.readFile(*file)
	}
	if folder, ok := msg.(*Folder); ok {
		r.folders <- *folder
		return nil
	}
	if sfAuth, ok := msg.(*SfAuth); ok {
		r.sfAuth <- *sfAuth
		return nil
	}
	return fmt.Errorf("Unexpected message type")
}

func (r *receiver) readFile(file File) error {
	c := make(chan []byte) // buffer by num blocks
	file.Content = c
	r.files <- file
	err := r.readContent(file.Size, c)
	if err != nil {
		return fmt.Errorf("Failed to read content for file %v %v: %v", file.ID, file.Name, err)
	}
	// read checksum
	return nil
}

func (r *receiver) readMsg(msgTypes ...msgIn) (msgIn, error) {
	bytes, err := r.readMsgBytes()
	if err != nil {
		return nil, fmt.Errorf("Failed to read message: %v", err)
	}
	fmt.Printf("raw: %v\n", string(bytes))
	for _, msg := range msgTypes {
		err = json.Unmarshal(bytes, msg)
		if err != nil || !msg.valid() {
			continue
		}
		return msg, nil
	}
	return nil, fmt.Errorf("Failed to parse message: %v", string(bytes))
}

func (r *receiver) readMsgBytes() ([]byte, error) {
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

func (r *receiver) readMsgSize() (uint16, error) {
	var size uint16
	err := binary.Read(r.rdr, binary.BigEndian, &size)
	return size, err
}

func (r *receiver) readContent(n uint64, c chan<- []byte) error {
	defer close(c)
	if n == 0 {
		return nil
	}
	var read uint64
	for read < n {
		var toRead uint64 = 4
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

type msgIn interface {
	valid() bool
}

func (s SfAuth) valid() bool {
	return s.T == "SfAuth"
}

func (f File) valid() bool {
	return f.T == "File"
}

func (f Folder) valid() bool {
	return f.T == "Folder"
}

func (r *receiver) SfAuth() <-chan SfAuth {
	return r.sfAuth
}

func (r *receiver) Files() <-chan File {
	return r.files
}

func (r *receiver) Folders() <-chan Folder {
	return r.folders
}
