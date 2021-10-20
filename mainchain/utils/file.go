package utils

import (
	"bufio"
	"errors"
	"io"
	"os"
)

func ReadFileToBytes(path string) ([]byte, error) {
	if !DirectoryOrFileExists(path) {
		return nil, errors.New(" file not exists.")
	}
	payload := make([]byte, 0)
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		bytes, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}else if err != nil {
			return nil, err
		}
		payload = append(payload, bytes...)
	}
	return payload, nil
}

func SaveBytesToFile(path string, payload []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	const Size = 4 * 1024
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	for i := 0; i < len(payload)/Size; i++ {
		if (i+1)*Size > len(payload) {
			if _, err := writer.Write(payload[i*Size:]); err != nil {
				return err
			}
			break
		}else {
			if _, err := writer.Write(payload[i*Size:(i+1)*Size]); err != nil {
				return err
			}
		}
	}
	return nil
}