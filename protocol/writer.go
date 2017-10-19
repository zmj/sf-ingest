package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

func NewWriter(wr io.Writer) Writer {
	return &writer{
		wr: wr,
	}
}

type writer struct {
	wr io.Writer
}

func (w *writer) writeMsg(msg interface{}) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Failed to serialize message: %v", err)
	}
	err = w.writeMsgSize(len(bytes))
	if err != nil {
		return fmt.Errorf("Failed to write message size: %v", err)
	}
	_, err = w.wr.Write(bytes)
	if err != nil {
		return fmt.Errorf("Failed to write message content: %v", err)
	}
	return nil
}

func (w *writer) writeMsgSize(size int) error {
	if size > math.MaxUint16 {
		return fmt.Errorf("Message size %v > max %v", size, math.MaxUint16)
	}
	return binary.Write(w.wr, binary.BigEndian, uint16(size))
}

func (w *writer) ItemDone(itemDone ItemDone) error {
	return w.writeMsg(itemDone)
}

func (w *writer) Error(err Error) error {
	return w.writeMsg(err)
}
