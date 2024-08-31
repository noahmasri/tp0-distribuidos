package common

import (
	"bytes"
	"encoding/binary"
)

type Bet struct {
	Name        string
	Surname		string
	ID			uint32
	Birthdate	string
	Number		uint16
}

func (bet *Bet) EncodeToBytes() []byte {
	var buffer bytes.Buffer

	binary.Write(&buffer, binary.LittleEndian, uint8(len(bet.Name)))
	binary.Write(&buffer, binary.LittleEndian, []byte(bet.Name))

	binary.Write(&buffer, binary.LittleEndian, uint8(len(bet.Surname)))
	binary.Write(&buffer, binary.LittleEndian, []byte(bet.Surname))

	binary.Write(&buffer, binary.LittleEndian, bet.ID)

	binary.Write(&buffer, binary.LittleEndian, []byte(bet.Birthdate))
	binary.Write(&buffer, binary.LittleEndian, bet.Number)

	return buffer.Bytes()
}