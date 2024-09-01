package common

import (
	"fmt"
	"os"
    "strings"
    "io"
)

type BetGetter struct {
	file        *os.File
    batchSize   int
    pending     []Bet // leftovers from a read from file operation
}

const FileReadBytes = 4096 // arbitrarilly choose to read 4096 bytes from file each time
const LineBreak = "\r\n"
const Separator = ","

func NewBetGetter(cliId string, batchSize int) *BetGetter {
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

func (bg *BetGetter) ReadBetsFromFile() ([]Bet, error){
    buf := make([]byte, FileReadBytes)
    read, err := bg.file.Read(buf)
    bets := []Bet{}

    if err != nil || read == 0 {
        return bets, err
    }

    // see where the last complete line ended
    lastCompleteRecord := strings.LastIndex(string(buf), LineBreak)

    // get strings from upto where the last complete bet was found
    betsStrings := strings.Split(string(buf[:lastCompleteRecord]), LineBreak)

    // move pointer back two bytes further (\r\n) from where last complete record ended
    returnOffset := int64(lastCompleteRecord + 2 - read)
    bg.file.Seek(returnOffset, io.SeekCurrent)

    for _, record := range betsStrings {
        bet, betErr := BetFromCSV(strings.Split(record, Separator))
        if betErr != nil {
            log.Errorf("action: parse_record | result: fail | error: could not create bet from record | entry: %v",
                record,
            )
            continue
        }
        fmt.Printf("Bet: %+v\n", bet)
        bets = append(bets, bet)
    }
    fmt.Printf("got %v entrys\n", len(betsStrings))
    return bets, err
}

func Transfer(arr0 *[]Bet, arr1 *[]Bet) {
    *arr0, *arr1 = *arr1, *arr0
}

/*
func (bg *BetGetter) GetBatch() ([]Bet, error){
    bets := []Bet{}
    transfer(bg.pending, bets)
    fmt.Printf(bg.pending)
    fmt.Printf(bets)
}
*/