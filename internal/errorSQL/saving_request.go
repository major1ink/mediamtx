package errorsql

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (f *Filesqlerror) SavingRequest(dir, query string) error {
	data := time.Now().Format("2006-01-02")
	if f.File == nil {
		err := f.createFile(data, dir, query)
		if err != nil {
			return err
		}
	}
	if data != f.Data {
		f.File.Close()
		err := f.createFile(data, dir, query)
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

func (f *Filesqlerror) createFile(data, dir, query string) error {

	f.Data = data
	dir = fmt.Sprintf("%s/%s", dir, data)
	if strings.Contains(query, "pathStream") {
		dir = fmt.Sprintf("%s/%s", dir, "PathStream.txt")
	} else {
		dir = fmt.Sprintf("%s/%s", dir, "Stream.txt")
	}
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Dir(dir), os.ModePerm)
			if err != nil {
				return err 
			}
			file, err := os.OpenFile(dir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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
	file, err := os.OpenFile(dir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	f.File = file
	return nil
}
