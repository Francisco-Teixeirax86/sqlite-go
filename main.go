package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func main() {
	databaseFilePath := os.Args[1]
	command := os.Args[2]

	switch command {
	case ".dbinfo":
		databaseFile, err := os.Open(databaseFilePath)
		if err != nil {
			log.Fatal(err)
		}

		header := make([]byte, 100) //first 100 bytes are reserved for the header
		_, err = databaseFile.Read(header)
		if err != nil {
			log.Fatal(err)
		}

		var pageSize uint16
		if err = binary.Read(bytes.NewBuffer(header[16:18]), binary.BigEndian, &pageSize); err != nil {
			fmt.Println("Failed to read integer:", err)
			return
		}

		leaf := make([]byte, pageSize-100)
		_, err = databaseFile.Read(leaf)
		if err != nil {
			log.Fatal(err)
		}

		var numTables uint16
		if err = binary.Read(bytes.NewBuffer(leaf[3:5]), binary.BigEndian, &numTables); err != nil {
			fmt.Println("Failed to read integer:", err)
			return
		}

		fmt.Println("PageSize: ", pageSize)
		fmt.Println("NumTables: ", numTables)
	default:
		fmt.Println("Unsupported command:", command)
		os.Exit(1)
	}

}
