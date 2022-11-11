package lib

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

func FileExists(path string) (bool, error) {
	var exists bool = false
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		exists = false
	} else if err != nil {
		return false, fmt.Errorf("%w", err)
	} else {
		exists = true
	}
	return exists, nil
}

// These functions work with arrays of JSON objects. OpenJsonFile
// expects that the file does not exist or is a valid JSON array.
// That means that an empty file will throw an error.
func OpenJsonFile(path string) (*os.File, error) {
	existed, err := FileExists(path)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	// If the file did not exist, we just add the start of the array
	// and are done.
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
		// We seek the file to find where the array ends.
		cursor -= 1
		f.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		f.Read(char)

		if char[0] == ']' {
			// After we have found the end of the array we need to seek to figure out
			// if we need to add a comma or not.
			var inner_cur int64 = 0
			for {
				inner_cur -= 1
				f.Seek(inner_cur+cursor, io.SeekEnd)

				char2 := make([]byte, 1)
				f.Read(char2)
				if char2[0] == '}' {
					f.WriteAt([]byte{','}, fsize+cursor)
					return f, nil
				}
				if cursor+inner_cur == -fsize {
					// If we reach the end of file and have not found the end of a JSON object,
					// we can expect that the file is an empty array.
					break
				}
			}

			f.WriteAt([]byte{' '}, fsize+cursor)
			return f, nil
		}

		// The previous process may have died unexpectedly and not closed the file.
		if char[0] == ',' {
			return f, nil
		}

		if cursor == -fsize { // stop if we are at the beginning
			return f, fmt.Errorf("end of file reached without finding the end of a JSON array")
		}
	}
}

func CloseJsonFile(f *os.File) {
	var cursor int64 = 0
	stat, _ := f.Stat()
	fsize := stat.Size()

	for {
		cursor -= 1
		f.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		f.Read(char)

		if char[0] == ',' || char[0] == '[' || char[0] == '}' {
			if char[0] == ',' {
				// We need to write over the comma.
				_, err := f.WriteAt([]byte("]"), fsize+cursor)
				if err != nil {
					log.Printf("could not write file")
				}
			} else if char[0] == '[' || char[0] == '}' {
				// We don't want to write over the start of the array.
				_, err := f.WriteAt([]byte("]"), fsize+cursor+1)
				if err != nil {
					log.Printf("could not write file")
				}
			}
			break
		}

		if cursor == -fsize { // stop if we are at the beginning
			log.Printf("end of file reached without finding the last JSON object or start of array")
			break
		}
	}
	err := f.Close()
	if err != nil {
		log.Printf("closed with errors: %v", err)
	}
}

func WriteJsonObjectToFile(f *os.File, obj []byte) error {
	f.Seek(0, io.SeekEnd)

	if _, err := f.Write(obj); err != nil {
		return err
	}
	if _, err := f.Write([]byte(",")); err != nil {
		return err
	}
	return nil
}
