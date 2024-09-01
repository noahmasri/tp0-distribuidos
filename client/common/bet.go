package common

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

type Bet struct {
	Name        string
	Surname		string
	ID			uint32
	Birthdate	string
	Number		uint16
}

func BetFromCSV(betRecord []string) (Bet, error) {
    var bet Bet
    id, err := strconv.ParseUint(betRecord[2], 10, 32)
    if err != nil {
        return bet, err
    }

    number, err := strconv.ParseUint(betRecord[4], 10, 16)
    if err != nil {
        return bet, err
    }

    bet = Bet{
        Name:      betRecord[0],
        Surname:   betRecord[1],
        ID:        uint32(id),
        Birthdate: betRecord[3],
        Number:    uint16(number),
    }

    return bet, nil
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