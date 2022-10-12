package lib

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func FileExists(path string) (bool, error) {
	var exists bool = false
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		exists = false
	} else if err != nil {
		return false, fmt.Errorf("%v", err)
	} else {
		exists = true
	}
	return exists, nil
}
func OpenJsonFile(path string) (*os.File, error) {
	existed, err := FileExists(path)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	if !existed {
		_, err = f.WriteString("[\n")
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	var cursor int64 = 0
	stat, _ := f.Stat()
	fsize := stat.Size()

	for {
		cursor -= 1
		f.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		f.Read(char)

		if cursor != -1 && (char[0] == ']' || char[0] == '}') { // stop if we find a JSON object or array
			if char[0] == '}' {
				return f, fmt.Errorf("could not find JSON array end")
			}
			f.Write([]byte{' '})
		}

		if cursor == -fsize { // stop if we are at the beginning
			return f, fmt.Errorf("could not find JSON array end")
		}
	}

	return f, nil
}

func CloseJsonFile(f *os.File) error {
	var cursor int64 = 0
	stat, _ := f.Stat()
	fsize := stat.Size()
	var err1 error

	for {
		cursor -= 1
		f.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		f.Read(char)

		if cursor != -1 && (char[0] == '}' || char[0] == ',') { // stop if we find a JSON object or array
			if char[0] == ',' {
				if _, err := f.Write([]byte("\n]\n")); err != nil {
					err1 = err
				}
			} else {
				err1 = fmt.Errorf("could not find JSON array end")
			}
		}

		if cursor == -fsize { // stop if we are at the beginning
			err1 = fmt.Errorf("could not find JSON array end")
		}
	}
	err2 := f.Close()
	if err2 != nil || err1 != nil {
		return fmt.Errorf("closed with errors: %v, %v", err1, err2)
	}
	return nil
}

func WriteJsonObjectToFile(f *os.File, obj []byte) error {
	if _, err := f.Write(obj); err != nil {
		return err
	}
	if _, err := f.Write([]byte(",")); err != nil {
		return err
	}
}
