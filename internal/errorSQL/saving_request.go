package errorsql

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (f *Filesqlerror) SavingRequest(filename, query string) error {
	data := time.Now().Format("2006-01-02")
	if f.File == nil {
		err := f.createFile(data, filename, query)
		if err != nil {
			return err
		}
	}
	if data != f.Data {
		f.File.Close()
		err := f.createFile(data, filename, query)
		if err != nil {
			return err
		}
	}

	message := []byte(fmt.Sprintf("%s,\n", strings.Split(query, "VALUES")[1]))
	_, err := f.File.Write(message)
	if err != nil {
		return err
	}
	return nil
}

func (f *Filesqlerror) createFile(data, filename, query string) error {

	f.Data = data
	filename = fmt.Sprintf("%s_%s.txt", strings.Split(filename, ".txt")[0], data)
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Dir(filename), os.ModePerm)
			if err != nil {
				return err 
			}
			file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return err
			}
			f.File = file
			insert := []byte(fmt.Sprintf("%sVALUES\n", strings.Split(query, "VALUES")[0]))
			_, err = f.File.Write(insert)
			if err != nil {
				return err
			}
			return nil

		} else {
			return err
		}
	}
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	f.File = file
	return nil
}
