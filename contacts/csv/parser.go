package csv

import (
	"encoding/csv"
	"os"
)

func ParseCsv(fileName string) ([]string, [][]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	return header, records, nil
}
