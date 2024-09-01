package common

import (
	"fmt"
	"os"
    "strings"
    "io"
)

type BetGetter struct {
	file            *os.File
    batchSize       int
    lastBatchSize   int
    pending         []Bet // leftovers from a read from file operation
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

	betGetter := &BetGetter{
		file: file,
		batchSize: batchSize,
        lastBatchSize: 0,
	}
	return betGetter
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
        bets = append(bets, bet)
    }
    return bets, err
}

func Transfer(arr0 *[]Bet, arr1 *[]Bet) {
    *arr0, *arr1 = *arr1, *arr0
}

// returns the amount of bets that were not inserted into 'bets' slice
func AppendMissing(missing *int, bets *[]Bet, readBets []Bet) int {
    if *missing > len(readBets){
        *bets = append(*bets, readBets...)
        *missing -= len(readBets)
        return 0
    }

    // got more or same bets than what i needed, no more missing
    extra := len(readBets) - *missing
    *bets = append(*bets, readBets[:len(readBets) - extra]...) 
    *missing = 0
    return extra
}

func (bg *BetGetter) GetBatch() ([]Bet, error){
    bets := []Bet{}
    Transfer(&bg.pending, &bets)
    missing := bg.batchSize - len(bets)

    for missing > 0 {
        readBets, err := bg.ReadBetsFromFile()
        if err != nil {
            if err.Error() == "EOF" {
                break
            }
            return bets, err 
        }
        extra := AppendMissing(&missing, &bets, readBets)
        if extra > 0 {
            bg.pending = append(bg.pending, readBets[len(readBets) - extra:]...)
        }
    }
    bg.lastBatchSize = len(bets)
    return bets, nil
}

func (bg *BetGetter) ReadEntireFileInBatches(){
    acumulado := 0
	for {
        bets, err := bg.GetBatch()
        if err != nil {
            fmt.Println("Error:", err)
            break
        }
        if len(bets) == 0 {
            fmt.Println("Nothing read, reached EOF")
            break
        }
        
		fmt.Printf("got %v bets from batch\n", len(bets))
		acumulado += len(bets)
		fmt.Printf("acumulado upto now %v\n", acumulado)
    }

	fmt.Printf("got %v bets from whole file\n", acumulado)

}