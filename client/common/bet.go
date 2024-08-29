package common

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"fmt"
)

type Bet struct {
	Name        string
	Surname		string
	ID			uint32
	Date		string
	Number		uint16
}

func (bet *Bet) EncodeToBytes() []byte {
	var buffer bytes.Buffer

	binary.Write(&buffer, binary.LittleEndian, uint8(len(bet.Name)))
	binary.Write(&buffer, binary.LittleEndian, []byte(bet.Name))

	binary.Write(&buffer, binary.LittleEndian, uint8(len(bet.Surname)))
	binary.Write(&buffer, binary.LittleEndian, []byte(bet.Surname))

	binary.Write(&buffer, binary.LittleEndian, bet.ID)

	binary.Write(&buffer, binary.LittleEndian, []byte(bet.Date))
	binary.Write(&buffer, binary.LittleEndian, bet.Number)

	return buffer.Bytes()
}

func EncodeAgencyData(cli_id string, bet Bet) []byte {
	agency, err := strconv.ParseInt(cli_id, 10, 8)
	if err != nil {
		log.Fatalf("action: parse_bets | result: fail | error: %v | msg: agency number is invalid", err)
		return nil
	}

	var buffer bytes.Buffer
	bet_bytes := bet.EncodeToBytes()
	binary.Write(&buffer, binary.LittleEndian, uint8(agency))
	binary.Write(&buffer, binary.LittleEndian, uint16(len(bet_bytes)))
	buffer.Write(bet_bytes)
	fmt.Printf("Encoded Bet: %v\n", buffer)
	return buffer.Bytes()
}