package common

import (
	"fmt"
	"os"
    "strings"
    "io"
)

type BetGetter struct {
	file        *os.File
    offset      int
    batchSize   uint8
}

const FileReadBytes = 4096 // arbitrarilly choose to read 4096 bytes from file each time
const LineBreak = "\r\n"
const Separator = ","

func NewBetGetter(cliId string, batchSize uint8) *BetGetter {
    filePath := fmt.Sprintf("/.data/agency-%s.csv", cliId)
    file, err := os.Open(filePath)

    if err != nil { 
        log.Fatal("Error opening data file: ", err) 
        return nil
    }

	client := &BetGetter{
		file: file,
		batchSize: batchSize,
	}
	return client
}

func (bg *BetGetter) Read() ([]Bet, error){
    buf := make([]byte, FileReadBytes)
    read, err := bg.file.Read(buf)
    bets := []Bet{}

    if err != nil || read == 0{
        return bets, err
    }

    // see where the last complete line ended
    lastCompleteRecord := strings.LastIndex(string(buf), LineBreak)

    // get strings from upto where the last complete bet was found
    betsStrings := strings.Split(string(buf[:lastCompleteRecord]), LineBreak)

    // move pointer back two bytes further (\r\n) from where last complete record ended
    returnOffset := int64(lastCompleteRecord + 2 - read)
    bg.file.Seek(returnOffset, io.SeekCurrent)
    fmt.Println(betsStrings)
    fmt.Println(lastCompleteRecord)

    return bets, err
}