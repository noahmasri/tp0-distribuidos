package common

import (
	"encoding/csv"
	"fmt"
	"os"
)

type BetGetter struct {
	file        *os.File
    reader      *csv.Reader
    batchSize   uint8
}

func NewBetGetter(cliId string, batchSize uint8) *BetGetter {
    filePath := fmt.Sprintf("/.data/agency-%s.csv", cliId)
    file, err := os.Open(filePath)

    if err != nil { 
        log.Fatal("Error opening data file: ", err) 
        return nil
    }

	client := &BetGetter{
		file: file,
		reader: csv.NewReader(file),
		batchSize: batchSize,
	}
	return client
}

func (bg *BetGetter) Read(){
    record, err := bg.reader.Read()
    if err != nil { 
        fmt.Println("Error reading records") 
    } 
    
	fmt.Println(record)
    bet, betErr := BetFromCSV(record)
    if betErr != nil{
        fmt.Println("Error creating bet from record") 
    }
    fmt.Printf("Bet: %+v\n", bet)
}