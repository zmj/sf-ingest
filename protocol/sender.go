package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

func NewSender(wr io.Writer) Sender {
	return &sender{
		wr: wr,
	}
}

type sender struct {
	wr io.Writer
}

func (s *sender) writeMsg(msg interface{}) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Failed to serialize message: %v", err)
	}
	err = s.writeMsgSize(len(bytes))
	if err != nil {
		return fmt.Errorf("Failed to write message size: %v", err)
	}
	_, err = s.wr.Write(bytes)
	if err != nil {
		return fmt.Errorf("Failed to write message content: %v", err)
	}
	return nil
}

func (s *sender) writeMsgSize(size int) error {
	if size > math.MaxUint16 {
		return fmt.Errorf("Message size %v > max %v", size, math.MaxUint16)
	}
	return binary.Write(s.wr, binary.BigEndian, uint16(size))
}

func (s *sender) ItemDone(itemDone ItemDone) error {
	return s.writeMsg(itemDone)
}

func (s *sender) ServerError(err ServerError) error {
	return s.writeMsg(err)
}
