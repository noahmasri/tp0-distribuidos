package common

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ReadAgencyBets(cli_id string){
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
}