package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func parseUInt16(file *os.File) uint16 {
	// Parse a 2-byte big-endian integer
	bytes := make([]byte, 2)
	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return binary.BigEndian.Uint16(bytes)
}

func parseVarint(file *os.File) uint64 {
	// Parse a variable-length integer
	var value uint64
	var shift uint
	for {
		byteVal := make([]byte, 1)
		_, err := file.Read(byteVal)
		if err != nil {
			log.Fatal(err)
		}
		value |= uint64(byteVal[0]&0x7F) << shift
		if byteVal[0]&0x80 == 0 {
			break
		}
		shift += 7
	}
	return value
}

func parseRecord(file *os.File, numColumns int) (values []string) {
	// Parse the record body and extract column values
	headerSize := parseVarint(file)
	header := make([]byte, headerSize-1)
	_, err := file.Read(header)
	if err != nil {
		log.Fatal(err)
	}

	values = make([]string, numColumns)
	for i := 0; i < numColumns; i++ {
		serialType := header[i]
		var value []byte
		if serialType >= 13 {
			size := (serialType - 13) / 2
			value = make([]byte, size)
			_, err := file.Read(value)
			if err != nil {
				log.Fatal(err)
			}
		}
		values[i] = string(value)
	}
	return values
}

func parsePageHeader(file *os.File) (numberOfCells uint16) {
	// Read page header and extract the number of cells
	header := make([]byte, 8)
	_, err := file.Read(header)
	if err != nil {
		log.Fatal(err)
	}
	return binary.BigEndian.Uint16(header[3:5]) // Extract the number of cells
}

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

		fmt.Printf("database page size: %v\n", pageSize)
		fmt.Printf("number of tables: %v\n", numTables)
	case ".tables":
		databaseFile, err := os.Open(databaseFilePath)
		if err != nil {
			log.Fatal(err)
		}

		_, err = databaseFile.Seek(100, io.SeekStart) //Skip the header of the file
		numberOfCells := parsePageHeader(databaseFile)

		// Read Cell Pointers
		cellPointers := make([]uint16, numberOfCells)
		for i := 0; i < int(numberOfCells); i++ {
			cellPointers[i] = parseUInt16(databaseFile)
		}

		var tables []string
		for _, cellPointer := range cellPointers {
			_, _ = databaseFile.Seek(int64(cellPointer), io.SeekStart)
			parseVarint(databaseFile) //Parse the size of the payload
			parseVarint(databaseFile) //Parse the Rowid
			record := parseRecord(databaseFile, 5)

			tableType := record[0]
			tableName := record[1]

			//filter out the sqlite_sequence name
			if tableType == "table" && tableName != "sqlite_sequence" {
				tables = append(tables, tableName)
			}
		}

		fmt.Printf("%s\n", strings.Join(tables, " "))
	default:
		fmt.Println("Unsupported command:", command)
		os.Exit(1)
	}

}
