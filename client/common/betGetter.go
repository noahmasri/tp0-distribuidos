package common

import (
	"encoding/csv"
	"fmt"
	"os"
)

type BetGetter struct {
	file        *os.File
    batchSize   uint8
}

func InitBetGetter(cli_id string){
	filePath := fmt.Sprintf("/.data/agency-%s.csv", cli_id)
    file, err := os.Open(filePath)

    if err != nil { 
        log.Fatal("Error while reading the file", err) 
    }

    defer file.Close()

    reader := csv.NewReader(file)
	record, err := reader.Read()
    if err != nil { 
        fmt.Println("Error reading records") 
    } 
    
	fmt.Println(record)
    bet, betErr := BetFromCSV(record)
    if betErr != nil{
        fmt.Println("Error creating bet from record") 
    }
    fmt.Printf("Bet: %+v\n", bet)

    record, err = reader.Read()
    if err != nil { 
        fmt.Println("Error reading records") 
    } 
    fmt.Println(record)
}